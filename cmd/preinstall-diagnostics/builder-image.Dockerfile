FROM fedora

RUN dnf install -y make
RUN curl -Lo go.tar.gz https://go.dev/dl/go1.17.6.linux-amd64.tar.gz
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf go.tar.gz
ENV PATH=$PATH:/usr/local/go/bin
RUN go version