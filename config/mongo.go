package config

import (
	mgo "gopkg.in/mgo.v2"
)

// Connect - Connecting to mongodb
func Connect() (*mgo.Session, error) {
	session, err := mgo.Dial("159.89.193.168")
	if err != nil {
		return nil, err
	}

	return session, nil
}
