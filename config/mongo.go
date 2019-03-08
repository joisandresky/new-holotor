package config

import (
	mgo "gopkg.in/mgo.v2"
)

// Connect - Connecting to mongodb
func Connect() (*mgo.Session, error) {
	session, err := mgo.Dial("host.docker.internal")
	if err != nil {
		return nil, err
	}

	return session, nil
}
