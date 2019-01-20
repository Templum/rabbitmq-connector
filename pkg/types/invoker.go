package types

// Invoker is the Interface used by the OpenFaaS Connector SDK to perform invocations
// of Lambdas based on a provided topic and message
type Invoker interface {
	Invoke(topic string, message *[]byte)
}
