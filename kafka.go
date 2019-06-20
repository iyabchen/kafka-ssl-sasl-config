package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"strings"

	"github.com/Shopify/sarama"
)

func main() {
	var (
		brokers    string
		saslEnable bool
		username   string
		password   string
		tlsEnable  bool
		clientcert string
		clientkey  string
		cacert     string
	)

	flag.StringVar(&brokers, "brokers", "localhost:9092", "Common separated kafka hosts")

	flag.BoolVar(&saslEnable, "sasl", true, "SASL enable")
	flag.StringVar(&username, "username", "admin", "SASL Username")
	flag.StringVar(&password, "password", "admin", "SASL Password")

	flag.BoolVar(&tlsEnable, "tls", true, "TLS enable")
	flag.StringVar(&clientcert, "cert", "/home/ichen/kafka-ssl-sasl-config/certs/kafka.client.pem", "Client Certificate")
	flag.StringVar(&clientkey, "key", "/home/ichen/kafka-ssl-sasl-config/certs/kafka.client.nokey", "Client Key")
	flag.StringVar(&cacert, "ca", "/home/ichen/kafka-ssl-sasl-config/certs/ca-cert", "CA Certificate")
	flag.Parse()

	config := sarama.NewConfig()
	if saslEnable {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = "admin"
		config.Net.SASL.Password = "admin"
	}

	if tlsEnable {

		// load client cert
		clientcert, err := tls.LoadX509KeyPair(clientcert, clientkey)
		if err != nil {
			log.Fatal(err)
		}

		// load ca cert pool
		cacert, err := ioutil.ReadFile(cacert)
		if err != nil {
			log.Fatal(err)
		}
		cacertpool := x509.NewCertPool()
		cacertpool.AppendCertsFromPEM(cacert)

		// generate tlcconfig
		tlsConfig := tls.Config{}
		tlsConfig.RootCAs = cacertpool
		tlsConfig.Certificates = []tls.Certificate{clientcert}
		tlsConfig.BuildNameToCertificate()
		tlsConfig.InsecureSkipVerify = true // This can be used on test server if domain does not match cert:
		config.Net.TLS.Enable = true

		config.Net.TLS.Config = &tlsConfig

	}

	config.Producer.Return.Successes = true
	client, err := sarama.NewClient(strings.Split(brokers, ","), config)
	if err != nil {
		log.Fatal("Unable to create kafka client " + err.Error())
	}

	err = client.RefreshMetadata()
	if err != nil {
		log.Fatal(err)
	}

	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()

	producer.SendMessage(&sarama.ProducerMessage{Topic: "test", Key: nil, Value: sarama.StringEncoder("message")})

}
