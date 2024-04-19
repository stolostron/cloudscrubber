FROM registry.ci.openshift.org/stolostron/builder:go1.21-linux

RUN mkdir /cloud

WORKDIR /cloud

# copy source code into the container
COPY . .

# build the go application
RUN go build -o cloud .

# set environment variable
ENV cloudtask="awstag"

# run the executable
CMD ["./cloud"]