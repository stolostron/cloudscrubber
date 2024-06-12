FROM registry.ci.openshift.org/stolostron/builder:go1.21-linux

RUN mkdir /cloud

WORKDIR /cloud

# copy source code into the container
COPY . .

# build the go application
RUN go build -o cloud .

# set cloud task environment variable
# ENV CLOUD_TASK="awsextend"

# run the executable
CMD ["./cloud"]