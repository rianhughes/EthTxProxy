# Download base image ubuntu 22.04
FROM ubuntu:20.04

# Install git and C-compiler and latest go needed for geth
RUN apt-get -y update && \
    apt-get -y install build-essential 

RUN apt-get -y install wget && \
    wget  https://go.dev/dl/go1.20.2.linux-amd64.tar.gz && \
    tar -xvf go1.20.2.linux-amd64.tar.gz && \
    mv go /usr/local  


# Configure Go
ENV GOROOT /usr/local/go 
ENV GOPATH /go 
ENV PATH $GOROOT/bin:$PATH

# Install pip3 with python3.8
RUN apt-get -y install python3 \
    python3-pip


# install nodejs v14
RUN apt update && apt -y install curl dirmngr apt-transport-https lsb-release ca-certificates
RUN curl -sL https://deb.nodesource.com/setup_14.x | bash - 
# npm for ganache-cli install
RUN apt-get -y install nodejs \
    yarn 

# install ganache-cli v7.8.0
RUN npm install -g ganache

RUN apt-get install -y git
RUN git clone https://github.com/rianhughes/EthTxProxy_ReverseProxy

# Add the current directory to the container
RUN cd EthTxProxy_ReverseProxy && \
    go mod tidy && \ 
    python3 -m pip install -r requirements.txt



# Expose ports and entrypoint
#EXPOSE 8545 8546 30303 30303/udp
#ENTRYPOINT ["geth"]
#ENTRYPOINT ["ganache-cli", "-h", "0.0.0.0", "-p", "8545", "&&", "go", "run", "main.go"]

#https://hub.docker.com/r/trufflesuite/ganache-cli/
#https://levelup.gitconnected.com/run-the-ganache-cli-inside-the-docker-container-5e70bc962bfe
