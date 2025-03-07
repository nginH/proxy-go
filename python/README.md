


**Components Interaction:**

1.  **Web Request Arrival:** A web request arrives at the Go Proxy Server.
2.  **Cache Lookup (Go Proxy):** The Go proxy first checks Redis for a cached response.
3.  **Cache Decision Request (Go Proxy -> Python RL Service):** If no cache hit, the Go proxy sends a `GetCacheDecision` gRPC request to the Python RL Service, providing the request method and path.
4.  **RL Decision Making (Python RL Service):**
    *   The Python RL Service extracts features for the requested route from the `RequestLog`.
    *   The RL Agent uses its trained policy to determine the optimal caching action (Don't cache, Cache with short/medium/long TTL).
    *   The decision (should\_cache, ttl, confidence) is sent back to the Go Proxy via gRPC.
5.  **Caching (Go Proxy):** Based on the RL decision, the Go Proxy either caches the response in Redis with the recommended TTL or fetches the response from the origin server without caching.
6.  **Response to Client (Go Proxy):** The Go Proxy returns the response to the client.
7.  **Request Logging (Go Proxy -> Python RL Service):**  After handling the request (regardless of cache hit or miss), the Go Proxy sends a `LogRequest` gRPC call to the Python RL Service, providing detailed request information (method, path, response time, cache hit status, etc.). This log data is used for training the RL agent.
8.  **Background Training (Python RL Service):** The `RLTrainer` thread in the Python RL Service periodically trains the RL agent using the accumulated request logs to improve its caching policy over time.

## Detailed Explanation of Python RL Service Components

### 1. Request Log (`RequestLog` Class)

The `RequestLog` class is responsible for:

*   **Storing Web Request Data:** It maintains a list of recent web requests, limited by `max_log_size`. Each request is stored as a dictionary containing information like method, path, parameters, query, IP, timestamp, response time, cache hit status, and status code.
*   **Aggregating Route Statistics:** It calculates and stores statistics for each unique route (method + path combination), including:
    *   `count`: Total number of requests for the route.
    *   `avg_time`: Average response time for the route.
    *   `hit_rate`: Cache hit rate for the route.
    *   `miss_rate`: Cache miss rate for the route.
    *   `last_accessed`: Timestamp of the last access.
    *   `access_frequency`: Number of requests per minute for the route.

**Key Methods:**

*   `add_request(request_data: Dict)`: Adds a new request to the log and updates route statistics.
*   `get_route_features(method: str, path: str) -> Dict`: Extracts features for a given route, which are used as the state in the RL environment. These features include `count`, `avg_time`, `hit_rate`, `miss_rate`, `access_frequency`, and `recency`.
*   `get_popular_routes(threshold=10) -> List[Dict]`: Returns a list of popular routes based on a request count threshold.
*   `get_all_routes() -> List[Tuple[str, str]]`: Returns a list of all unique routes logged.

### 2. Redis Cache (`RedisCache` Class)

The `RedisCache` class provides a simple interface for interacting with Redis to perform caching operations:

*   `get_cached_response(method: str, path: str) -> Optional[dict]`: Retrieves a cached response from Redis for a given route.
*   `set_cached_response(method: str, path: str, data: Any, ttl: int) -> bool`: Caches a response in Redis with a specified Time-To-Live (TTL).
*   `invalidate_cache(method: str, path: str) -> bool`: Invalidates (deletes) the cache entry for a route in Redis.
*   `get_all_cached_keys() -> List[str]`: Retrieves all cache keys currently stored in Redis (with a prefix).

### 3. Cache Environment (`CacheEnvironment` Class)

The `CacheEnvironment` class defines the Reinforcement Learning environment where the agent learns to make caching decisions. It is a `gym.Env` class, adhering to the OpenAI Gym interface.

**Key Components of the Environment:**

*   **State Space (Observation Space):**  The state represents the current situation for a given route, based on features extracted from the `RequestLog`. The observation space is a 6-dimensional continuous space defined by `gym.spaces.Box`:

    *   `[request_count, avg_response_time, hit_rate, miss_rate, access_frequency, recency]`

    *   **Recency Transformation:** Recency is transformed to a value between 0 and 1 using the formula:

        ```
        recency_feature = 1.0 / (1.0 + recency)
        ```
        where `recency` is the time in seconds since the route was last accessed. This transformation ensures that more recent accesses result in higher recency feature values (closer to 1).

*   **Action Space:** The action space is discrete, representing different caching actions:

    *   `0`: **Don't cache.**
    *   `1`: **Cache with short TTL (1 minute).**
    *   `2`: **Cache with medium TTL (5 minutes).**
    *   `3`: **Cache with long TTL (15 minutes).**
    *   `4`: **Cache with very long TTL (1 hour).**

    *   The corresponding TTL values are stored in `self.ttl_values = [0, 60, 300, 900, 3600]` seconds.

*   **Reward Function:** The reward function is crucial for guiding the RL agent's learning. It is designed to incentivize actions that lead to better caching performance. The reward is calculated in the `step(action)` method based on the following components:

    *   **Hit Rate Component:**  Higher hit rate is desirable.
        ```
        hit_rate_component = features['hit_rate'] * 10
        ```

    *   **Frequency Component:** Higher access frequency makes caching more beneficial.
        ```
        freq_component = min(5, features['access_frequency'])
        ```
        The `min(5, ...)` clips the frequency component to a maximum value of 5 to prevent excessively high frequencies from dominating the reward.

    *   **Response Time Component:** Higher average response time indicates more time saved by caching.
        ```
        time_component = min(5, features['avg_time'] / 100)
        ```
        The average response time is divided by 100 and clipped to a maximum of 5, scaling the response time to a reasonable reward range.

    *   **Recency Component:** Caching recently accessed routes is more effective.
        ```
        recency_component = 5 * np.exp(-recency / 3600)
        ```
        This component uses an exponential decay function, giving higher rewards for more recent accesses. The decay is such that the reward decreases significantly after about an hour (3600 seconds).

    *   **TTL Penalty:** Penalizes using long TTLs for routes that are not frequently accessed, as this can lead to stale cache entries.
        ```
        ttl_penalty = -0.1 * (ttl / 60) * (1 - min(1, features['access_frequency']))
        ```
        The penalty is proportional to the TTL duration and inversely proportional to the access frequency.

    *   **Reward for Actions:**
        *   **Action 0 (Don't Cache):**
            ```
            reward = 1 - hit_rate_component - freq_component - time_component + ttl_penalty - 1
            ```
            This action gets a slightly negative base reward (-1) and subtracts the components that favor caching, effectively rewarding "not caching" when the route is not suitable for caching (low hit rate, frequency, response time).
        *   **Actions 1-4 (Cache with TTL):**
            ```
            reward = hit_rate_component + freq_component + time_component + recency_component + ttl_penalty
            ```
            These actions get positive rewards based on the components that indicate the benefits of caching.

*   **`reset()` method:** Resets the environment by randomly selecting a route from the `RequestLog` and returning its state.
*   `step(action: int) -> Tuple[np.ndarray, float, bool, Dict]`: Takes an action, calculates the reward based on the chosen action and route features, and returns the next state, reward, `done` flag (always `False` in this environment), and an `info` dictionary containing debugging information.

### 4. RL Agent (`RLCacheAgent` Class) and Policy Network (`Policy` Class)

*   **Policy Network (`Policy` Class):** This is a simple feedforward neural network that represents the RL agent's policy. It takes the state (route features) as input and outputs probabilities for each action in the action space.

    *   **Architecture:**
        *   Input Layer: Linear layer with `input_size` (6 in this case) input features and `hidden_size` (default 64) output features. ReLU activation.
        *   Hidden Layer: Linear layer with `hidden_size` input and output features. ReLU activation.
        *   Output Layer: Linear layer with `hidden_size` input features and `output_size` (5 in this case - number of actions) output features. Softmax activation to produce action probabilities.

    *   **Forward Pass:** The `forward(x)` method defines the network's forward pass, taking a state tensor `x` and returning a tensor of action probabilities.

*   **RL Agent (`RLCacheAgent` Class):** This class encapsulates the RL agent and its interaction with the environment and policy network.

    *   **Policy and Optimizer:** It initializes the `Policy` network and the Adam optimizer for training the policy.
    *   **Action Selection (`select_action(state)`):**
        *   Takes the current state as input.
        *   Passes the state through the `Policy` network to get action probabilities.
        *   Uses a `Categorical` distribution based on the probabilities to sample an action. This is a stochastic policy, allowing for exploration.
        *   Stores the log probability of the selected action (`log_probs`) and entropy (`entropies`) for policy update.

    *   **Action Probability Retrieval (`get_action_probs(state)`):**  Returns the action probabilities for a given state without sampling an action, used for making deterministic decisions in `GetCacheDecision` gRPC method (with epsilon-greedy exploration).

    *   **Reward Storage (`store_reward(reward)`):** Stores the received reward in a list (`rewards`) for policy update.

    *   **Policy Update (`update_policy()`):** Implements the REINFORCE algorithm (policy gradient) to update the policy network based on collected experiences (state, action, reward sequences).

        *   **Returns Calculation:** Calculates discounted returns for each step in an episode using the formula:
            ```
            G_t = r_t + γ * G_{t+1}
            ```
            where:
            *   `G_t` is the discounted return at time step `t`.
            *   `r_t` is the reward received at time step `t`.
            *   `γ` (gamma) is the discount factor (default 0.99).

        *   **Returns Normalization:** Normalizes the returns to have zero mean and unit standard deviation, which helps stabilize training.

        *   **Policy Loss Calculation:** Calculates the policy loss using the formula:
            ```
            L = - Σ [log_prob(a_t | s_t) * G_t] - β * Σ [H(π(·|s_t))]
            ```
            where:
            *   `log_prob(a_t | s_t)` is the log probability of the action `a_t` taken in state `s_t` under the current policy.
            *   `G_t` is the discounted return for that step.
            *   `β` (beta) is the entropy bonus coefficient (default 0.01).
            *   `H(π(·|s_t))` is the entropy of the action distribution π(·|s_t) in state `s_t`. The entropy bonus encourages exploration by favoring policies that are less deterministic.

        *   **Gradient Descent:** Performs backpropagation to calculate gradients of the loss with respect to policy network parameters and updates the parameters using the Adam optimizer.

        *   **Memory Clearing:** Clears the stored `log_probs`, `rewards`, and `entropies` after policy update, preparing for the next batch of experiences.

    *   **Model Saving/Loading (`save_model()`, `load_model()`):** Methods to save and load the policy network's state dictionary to/from a file (`cache_policy.pt`), allowing for persistence and reuse of the trained policy.

### 5. gRPC Service (`CacheService` Class)

The `CacheService` class implements the gRPC server logic defined in `cache.proto`. It exposes the following gRPC methods:

*   `LogRequest(request, context) -> LogResponse`:
    *   Receives request log data from the Go proxy (`RequestLog` message).
    *   Adds the request to the `RequestLog`.
    *   Gets the state for the requested route.
    *   Selects an action using the RL agent's policy (for training data collection - exploration action).
    *   Returns a `LogResponse` with a recommended action and TTL.

*   `GetCacheDecision(request, context) -> CacheDecision`:
    *   Receives a `RouteRequest` (method, path) from the Go proxy.
    *   Gets the state for the requested route.
    *   Retrieves action probabilities from the RL agent (`get_action_probs`).
    *   Selects an action using epsilon-greedy exploration/exploitation strategy.
    *   Returns a `CacheDecision` message with `should_cache`, `ttl`, and `confidence`.

*   `InvalidateCache(request, context) -> InvalidateResponse`:
    *   Receives a `RouteRequest` for cache invalidation.
    *   Calls `redis_cache.invalidate_cache` to invalidate the cache entry in Redis.
    *   Returns an `InvalidateResponse` indicating success.

*   `GetPopularRoutes(request, context) -> PopularRoutesResponse`:
    *   Receives a `PopularRoutesRequest` with a threshold.
    *   Calls `request_log.get_popular_routes` to get popular routes based on the threshold.
    *   Returns a `PopularRoutesResponse` containing a list of `RouteStats` messages for popular routes.

### 6. RL Trainer (`RLTrainer` Class)

The `RLTrainer` class is a background thread responsible for periodically training the RL agent.

*   **Training Loop:** Runs in a separate thread and performs training in the background without blocking the gRPC server.
*   **Training Intervals:**
    *   `reward_interval`: Controls how frequently `_train_step` is called (default 10 seconds). `_train_step` performs a small training step based on recent experiences.
    *   `training_interval`: Controls how frequently `_train_epoch` is called (default 600 seconds). `_train_epoch` performs a full training epoch over multiple episodes, providing more stable policy updates.
*   **Training Steps (`_train_step`)**: Called more frequently, collects experiences for a few steps, and updates the policy based on these recent experiences.
*   **Training Epochs (`_train_epoch`)**: Called less frequently, performs a full training epoch over multiple episodes, providing more stable policy updates and potentially better exploration.
*   **Model Saving:** Periodically saves the trained policy model after each training epoch.
*   **Stopping the Trainer:** Provides a `stop()` method to gracefully stop the training thread.

## Example Scenario: Under the Hood ML

Let's consider a hypothetical scenario to illustrate how the RL agent makes a caching decision:

1.  **Request Arrives:** A `GET` request for `/api/products` arrives at the Go Proxy.

2.  **Cache Miss:** The Go Proxy checks Redis and finds no cached response for `GET /api/products`.

3.  **GetCacheDecision Request:** The Go Proxy sends a `GetCacheDecision` gRPC request to the Python RL Service for `method="GET", path="/api/products"`.

4.  **State Feature Extraction (Python Service):** The `CacheService` in Python retrieves features for `GET /api/products` from the `RequestLog`. Let's assume the features are:

    ```
    features = {
        'count': 500,
        'avg_time': 250,  // ms
        'hit_rate': 0.7,
        'miss_rate': 0.3,
        'access_frequency': 15, // requests per minute
        'recency': 60,      // seconds ago
    }
    ```

    The state vector passed to the RL agent would be:

    ```
    state = np.array([500, 250, 0.7, 0.3, 15, 1.0 / (1.0 + 60)])
    ```

5.  **Policy Network Inference (RL Agent):** The `RLCacheAgent` gets action probabilities from its `Policy` network for this `state`:

    ```
    action_probs = agent.get_action_probs(state)
    # Let's say action_probs is: [0.1, 0.15, 0.25, 0.3, 0.2]
    # These correspond to [Don't Cache, 1min TTL, 5min TTL, 15min TTL, 1hr TTL]
    ```

6.  **Epsilon-Greedy Action Selection:** With an exploration rate of `epsilon = 0.1`, the agent:
    *   With 10% probability, chooses a random action from [0, 1, 2, 3, 4].
    *   With 90% probability, chooses the action with the highest probability, which is action `3` (15min TTL) in this example (probability 0.3).

    Let's assume the agent chooses action `3` (15min TTL) in this case (either by exploration or exploitation).

7.  **Cache Decision Response:** The `CacheService` returns a `CacheDecision` gRPC response to the Go Proxy:

    ```
    CacheDecision {
        should_cache: true,
        ttl: 900,  // 15 minutes in seconds
        confidence: 0.3  // Probability of the chosen action
    }
    ```

8.  **Caching in Go Proxy:** The Go Proxy receives the `CacheDecision`, fetches the response from the origin server, and caches it in Redis with a TTL of 900 seconds.

9.  **Request Logging (Go Proxy):**  After serving the response, the Go Proxy sends a `LogRequest` gRPC call to the Python RL Service with details of the request, including the response time and `cache_hit=false`.

10. **Background Training (Python Service):** Over time, as more requests are logged, the `RLTrainer` will use this data to update the `Policy` network, adjusting the action probabilities for different states to improve the overall caching performance based on the defined reward function. For example, if the agent consistently chooses long TTLs for routes that become less frequent, the negative TTL penalty in the reward function will eventually encourage the agent to choose shorter TTLs or "Don't Cache" for such routes during training updates.

## Formulas Summary

*   **Recency Feature Transformation:**
    ```
    recency_feature = 1.0 / (1.0 + recency)
    ```

*   **Reward Function Components:**
    ```
    hit_rate_component = features['hit_rate'] * 10
    freq_component = min(5, features['access_frequency'])
    time_component = min(5, features['avg_time'] / 100)
    recency_component = 5 * np.exp(-recency / 3600)
    ttl_penalty = -0.1 * (ttl / 60) * (1 - min(1, features['access_frequency']))
    ```

*   **Reward for Actions:**
    *   **Action 0 (Don't Cache):**
        ```
        reward = 1 - hit_rate_component - freq_component - time_component + ttl_penalty - 1
        ```
    *   **Actions 1-4 (Cache with TTL):**
        ```
        reward = hit_rate_component + freq_component + time_component + recency_component + ttl_penalty
        ```

*   **Discounted Returns:**
    ```
    G_t = r_t + γ * G_{t+1}
    ```

*   **Policy Loss (REINFORCE with Entropy Bonus):**
    ```
    L = - Σ [log_prob(a_t | s_t) * G_t] - β * Σ [H(π(·|s_t))]
    ```

## Running the System

1.  **Install Python Dependencies:**
    ```bash
    pip install grpcio grpcio-tools redis gym torch numpy pandas
    ```

2.  **Generate gRPC Files:**
    ```bash
    python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. cache.proto
    ```
    (Ensure `cache.proto` is in the same directory).

3.  **Run Redis Server:** Make sure you have a Redis server running and accessible. Configure Redis connection details in `main()` function of `rl_cache_service.py` if needed.

4.  **Run the Python RL Service:**
    ```bash
    python rl_cache_service.py
    ```

5.  **Implement and Run Go Proxy Server:**  Refer to the `service` package in the provided Go code example. You need to complete the implementation of `fetchDataForRoute`, `getDefaultTTLForRoute`, and integrate the gRPC client calls (`GetCacheDecision`, `LogRequest`, etc.) into your Go proxy's request handling logic. Compile and run your Go proxy server.

6.  **Send Web Requests:** Send web requests to your Go proxy server. Observe the caching behavior, and monitor the logs of both the Go proxy and the Python RL service to understand the system's operation.

## Conclusion

This project demonstrates an intelligent web cache system that leverages Reinforcement Learning to dynamically optimize caching strategies. By monitoring web request patterns and learning from experience, the RL agent can make adaptive decisions on whether to cache routes and what TTLs to use, aiming to improve cache hit rates and reduce response latency. This approach offers a flexible and data-driven way to manage web caching in dynamic environments.

**Potential Improvements:**

*   **More Sophisticated Policy Network:** Explore deeper neural networks or recurrent networks for the policy.
*   **Advanced RL Algorithms:** Investigate more advanced RL algorithms like PPO, A2C, or DQN for potentially faster and more stable learning.
*   **Feature Engineering:** Explore more relevant features from web requests and system metrics to provide richer state information to the RL agent.
*   **Dynamic Reward Shaping:** Experiment with different reward functions to fine-tune the agent's learning behavior and optimize for specific performance metrics.
*   **Contextual Caching:** Incorporate contextual information (e.g., user agent, time of day) into the state to enable more context-aware caching decisions.
*   **Distributed RL Training:** Scale the RL training process for handling larger traffic volumes and more complex caching scenarios in a distributed environment.