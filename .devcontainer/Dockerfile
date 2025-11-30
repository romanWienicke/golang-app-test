FROM golang:latest

ARG GO_VERSION=go1.25.0
ARG GOLANG_CGI_LINT_VERSION=v2.4.0

# Create user vscode with UID 1000 and GID 1000, and set up home directory
RUN groupadd -g 1000 vscode && \
    useradd -m -s /bin/bash -u 1000 -g 1000 vscode     
# Copy files to /usr/local/bin/
COPY setup.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/setup.sh && \
    /usr/local/bin/setup.sh "${GO_VERSION}" "${GOLANG_CGI_LINT_VERSION}" && \
    rm /usr/local/bin/setup.sh   

  
USER vscode
