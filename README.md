# golang-app-test

## TLDR;
This is a repository using devcontainers for VSCode.  
It pre installs brew and act in the container.  
It uses docker-compose for setting up tests in docker.  

### Makefile
make tidy: run go tidy
make test: run tests
make run: start of the app
make pipeline-test: run act and test the pipeline including the app.
