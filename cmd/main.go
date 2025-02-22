package main

import (
	"os"

	logs "github.com/nginH/pkg/log"
)

func main() {
	logs.InitLogger()

	port := os.Getenv("PORT")
	if port == "" {
		logs.Error("PORT is not set")
		port := "6969"
		logs.Info("Setting PORT to default value: ", port)
	}

	// db := server.New()

	// logs.Info("Starting the reverse proxy server")
	// backendUrl, err := url.Parse("http://127.0.0.1:5001/psyched-span-426722-q0/us-central1/dev")
	// if err != nil {
	// 	panic(err)
	// }
	// proxy := httputil.NewSingleHostReverseProxy(backendUrl)

	// originalDirector := proxy.Director
	// proxy.Director = func(req *http.Request) {
	// 	originalDirector(req)
	// 	logs.Info("Request Host: ", req.Host)
	// 	logs.Info("Request URL: ", req.URL)
	// }
	// proxy.ModifyResponse = func(res *http.Response) error {
	// 	logs.Info("Response Status: ", res.Status)
	// 	return nil
	// }

	// proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
	// 	logs.Error("Error: ", err)
	// 	rw.WriteHeader(http.StatusBadGateway)
	// }

	// addr := "0.0.0.0:443"
	// logs.Info("Proxy server started at: https://", addr)
	// if err := http.ListenAndServeTLS(addr, "/Users/harshanand/proxy-go/cert.pem", "/Users/harshanand/proxy-go/key.pem", proxy); err != nil {
	// 	logs.Fatal(err)
	// }
}
