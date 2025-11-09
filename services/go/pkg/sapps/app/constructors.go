package app

import (
	"sapps/lib/connection"
	maindb "sapps/pkg/sapps/lib/db/main"
)

func provideDBConnections() []interface{} {
	return []interface{}{
		connection.InjectMainDB,
		connection.InjectFirebase,
		connection.NewChatGPT,
	}
}

func provideDBHandlers() []interface{} {
	return []interface{}{
		maindb.InjectMainDB,
	}
}

func httpAppConstructors() []interface{} {
	constructorsList := [][]interface{}{
		provideDBConnections(),
		provideDBHandlers(),
	}
	constructors := []interface{}{}
	for i := range constructorsList {
		constructors = append(constructors, constructorsList[i]...)
	}
	return constructors
}
