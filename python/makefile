venv:
	python3 -m venv venv

activate:
	. venv/bin/activate

run:
	. venv/bin/activate && python3 -m main.py

dumpRequirement:
	. venv/bin/activate && pip freeze > requirements.txt

install: venv
	. venv/bin/activate && pip install -r requirements.txt

clean:
	rm -rf venv
	find . -type d -name "__pycache__" -exec rm -rf {} +

.PHONY: venv activate run dumpRequirement install clean

