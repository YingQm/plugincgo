FROM ubuntu:16.04

RUN apt-get update

RUN apt-get install -y automake autoconf
RUN apt-get install -y gcc g++
RUN apt-get install -y libgmp-dev
RUN apt-get install -y libtool
RUN apt-get install -y flex bison
RUN apt-get install -y make
RUN apt-get install -y wget
RUN apt-get install -y git

RUN wget -P /tmp https://mirrors.ustc.edu.cn/golang/go1.10.2.linux-amd64.tar.gz

RUN tar -C /usr/local -xzf /tmp/go1.10.2.linux-amd64.tar.gz
RUN rm /tmp/go1.10.2.linux-amd64.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH" \
&& mkdir -p "$GOPATH/src/github.com/33cn/plugincgo"

WORKDIR $GOPATH/src/github.com/33cn/plugincgo
COPY . ./
RUN git clone https://github.com/33cn/libbyz-go.git $GOPATH/src/github.com/33cn/libbyz-go
ENV workspace $GOPATH/src/github.com/33cn/plugincgo/plugin/consensus/pbftlibbyz
ENV bftpath $GOPATH/src/github.com/33cn/libbyz-go

WORKDIR $bftpath/bft/sfslite-1.2
RUN autoreconf -i
RUN sh -x setup.gnu -f -i -s
RUN mkdir install
ENV SFSHOME $bftpath/bft/sfslite-1.2
RUN ./configure --prefix=$SFSHOME/install \
&& make CFLAGS="-Werror=strict-aliasing" CXXFLAGS="-fpermissive -DHAVE_GMP_CXX_OPS" \
&& make install

WORKDIR $bftpath/bft
RUN ln -s sfslite-1.2/install sfs \
&& ln -s /usr/lib gmp

WORKDIR $bftpath/bft/libbyz
RUN sed -i '418s/^.*$/  th_assert(sizeof(t.tv_sec) <= sizeof(long), "tv_sec is too big");/' Node.cc \
&& sed -i '420s/^.*$/  long int_bits = sizeof(long)*8;/' Node.cc \
&& make CPPFLAGS="-I../gmp -I../sfs/include/sfslite -g -Wall -DRECOVERY -fpermissive -DHAVE_GMP_CXX_OPS"

WORKDIR $workspace/test
RUN go build

ENTRYPOINT [ "./test" ]
