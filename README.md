# Kafka SASL SSL configuration

This repository records how to configure Kafka with SASL and SSL. SSL/TLS is a communication protocol for message encryption, whereas SASL is a framework for authentication.

The supported SASL mechanisms in Kafka includes GSSAPI (Kerberos), OAUTHBEARER, SCRAM, PLAIN. Note that PLAIN is not PLAINTEXT. PLAIN is using username and password for authentication.

Security protocols in Kafka.

```
PLAINTEXT - Un-authenticated, non-encrypted channel
SASL_PLAINTEXT - SASL authenticated, non-encrypted channel
SASL_SSL - SASL authenticated, SSL channel
SSL	- SSL channel
```

## SSL certs

Use the script `gen-ssl-certs.sh` to generate kafka server and client certs. Default password is **abcdefgh**.

1. Create a CA certificate

```
$ gen-ssl-certs.sh ca ca-cert.pem CN

// generates the following
ca-cert.pem  ca-cert.pem.key
```

2. Create server certificate for java

```
$ gen-ssl-certs.sh -k server ca-cert kafka. server

// generates the following
kafka.cert-file
ca-cert.srl
kafka.cert-signed
kafka.server.truststore.jks
kafka.server.keystore.jks

```

3. Create client certificates for java

```
$ gen-ssl-certs.sh -k client ca-cert kafka. client
// generate the following
kafka.client.keystore.jks  kafka.client.truststore.jks
```

4. Create client key

```
$ gen-ssl-certs.sh client ca-cert kafka. client
// generate the following (just ignore the error if there is any)
kafka.client.key
kafka.client.req
kafka.client.pem
```

5. Take the private key out of the client, for golang client

```
openssl rsa -in kafka.client.key -out kafka.client.nokey
// Prompts to key in the password, which is abcdefgh by default, and generates
kafka.client.nokey
```

## JAAS files for zookeeper and kafka

Create the following files for authentication in SASL, mechanism PLAIN. Note that the last line must have `;` as ending.

```
// zk-jaas.conf for zookeeper
Server{
	org.apache.kafka.common.security.plain.PlainLoginModule required
		user_admin="admin";
};
```

The name "Server" is the default name zookeeper checks.
`user_admin="admin"`, the `user_` prefix is taken out at parsing, meaning user name is `admin`, and its value is the password.

```
// kafka-jaas.conf for kafka
KafkaServer {
	org.apache.kafka.common.security.plain.PlainLoginModule required
		username="admin"
		password="admin"
		user_admin="admin";
};

KafkaClient {
	org.apache.kafka.common.security.plain.PlainLoginModule required
		username="admin"
		password="admin"
        user_admin="admin";
};

Client {
	org.apache.kafka.common.security.plain.PlainLoginModule required
		username="admin"
		password="admin"
        user_admin="admin";
};
```

KafkaServer is used for broker startup.

KafkaClient is used for client connection.

Client means Kafka as a client, connecting to zookeeper. If zookeeper does not have SASL config, then this part can be omitted.

username/password is for kafka inter broker init authentication when connected as a client.

## Client properties

For consumer and producer using Java. The first two lines are important for SASL.

```
// client.properties
security.protocol=SASL_SSL
sasl.mechanism=PLAIN
ssl.truststore.location=/certs/kafka.server.truststore.jks
ssl.truststore.password=abcdefgh
ssl.keystore.location=/certs/kafka.server.keystore.jks
ssl.keystore.password=abcdefgh
ssl.key.password=abcdefgh
```

## Start with docker-compose

The following setting is based on `wurstmeister/kafka` docker image, in this image java options can be passed in using KAFKA_OPTS. `docker-compose.yml` contains the details.

For zookeeper:

```
// zookeeper.properties, add the following line for SASL
authProvider.1=org.apache.zookeeper.server.auth.DigestAuthenticationProvider
jaasLoginRenew=3600000
requireClientAuthScheme=sasl
```

```
// add this to environment variable
KAFKA_OPTS: '-Djava.security.auth.login.config=/certs/zk-jaas.conf'
```

For kafka:

```
// environment variable that impact server.properties options

// listener and advertised listener cannot both use hostname as ip, else error shows the address is already binded.
// advertised listener must use an external IP in order to accept request out of docker container.
KAFKA_ADVERTISED_LISTENERS: 'SASL_SSL://${HOSTNAME}:9092'
KAFKA_LISTENERS: 'SASL_SSL://0.0.0.0:9092'
KAFKA_SECURITY_INTER_BROKER_PROTOCOL: 'SASL_SSL' // must be one of the listener protocol
KAFKA_SECURITY_PROTOCOL: 'SASL_SSL' // must be one of the listener protocol

// SSL config
KAFKA_SSL_KEYSTORE_LOCATION: '/certs/kafka.server.keystore.jks'
KAFKA_SSL_KEYSTORE_PASSWORD: 'abcdefgh'
KAFKA_SSL_KEY_PASSWORD: 'abcdefgh'
KAFKA_SSL_TRUSTSTORE_LOCATION: '/certs/kafka.server.truststore.jks'
KAFKA_SSL_TRUSTSTORE_PASSWORD: 'abcdefgh'
KAFKA_SSL_CLIENT_AUTH: 'required'

// SASL config
KAFKA_SASL_ENABLED_MECHANISMS: 'PLAIN'
KAFKA_SASL_MECHANISM_INTER_BROKER_PROTOCOL: 'PLAIN'
KAFKA_OPTS: '-Djava.security.auth.login.config=/certs/kafka-jaas.conf'
KAFKA_AUTHORIZER_CLASS_NAME: 'kafka.security.auth.SimpleAclAuthorizer'
KAFKA_SUPER_USERS: 'User:admin'
```

Start container with this command.

```
HOSTNAME=<host ip> docker-compose up -d
```

Inside the kafka container, producer and consumer can be run with these commands. They use client.properties, and the kafka-jaas for connection.

```
# pwd
/opt/kafka/bin

# env|grep KAFKA_OPTS
KAFKA_OPTS=-Djava.security.auth.login.config=/certs/kafka-jaas.conf

# ./kafka-console-producer.sh --topic test --broker-list localhost:9092 --producer.config /certs/client.properties

# ./kafka-console-consumer.sh --topic test --bootstrap-server localhost:9092 --consumer.config /certs/client.properties  --from-beginning

```

## Golang client connecting with SSL and SASL

Check `kafka.go`.

## Reference

- [SASL vs SSL](https://stackoverflow.com/questions/11347304/security-authentication-ssl-vs-sasl)

- https://docs.confluent.io/current/kafka/authentication_sasl/index.html

- [librdkafka](https://github.com/edenhill/librdkafka/wiki/Using-SSL-with-librdkafka)
