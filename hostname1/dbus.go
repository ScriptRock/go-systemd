// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Integration with the systemd hostnamed API.  See http://www.freedesktop.org/wiki/Software/systemd/hostnamed/
package hostname1

import (
	"fmt"
	"os"
	"strconv"

	"github.com/cloudhousetech/dbus"
)

const (
	dbusInterface = "org.freedesktop.hostname1"
	dbusPath      = "/org/freedesktop/hostname1"
)

// Conn is a connection to systemds dbus endpoint.
type Conn struct {
	conn   *dbus.Conn
	object dbus.BusObject
}

// New() establishes a connection to the system bus and authenticates.
func New() (*Conn, error) {
	c := new(Conn)

	if err := c.initConnection(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Conn) initConnection() error {
	var err error
	c.conn, err = dbus.SystemBusPrivate()
	if err != nil {
		return err
	}

	// Only use EXTERNAL method, and hardcode the uid (not username)
	// to avoid a username lookup (which requires a dynamically linked
	// libc)
	methods := []dbus.Auth{dbus.AuthExternal(strconv.Itoa(os.Getuid()))}

	err = c.conn.Auth(methods)
	if err != nil {
		c.conn.Close()
		return err
	}

	err = c.conn.Hello()
	if err != nil {
		c.conn.Close()
		return err
	}

	c.object = c.conn.Object("org.freedesktop.hostname1", dbus.ObjectPath(dbusPath))

	return nil
}

// close the connection to the dbus socket
func (c *Conn) Close() error {
	return c.conn.Close()
}

// GetProperties returns all hostname related properties
func (c *Conn) GetProperties() (map[string]interface{}, error) {
	var err error
	var props map[string]dbus.Variant

	err = c.object.Call("org.freedesktop.DBus.Properties.GetAll", 0, dbusInterface).Store(&props)
	if err != nil {
		return nil, err
	}

	out := make(map[string]interface{}, len(props))
	for k, v := range props {
		out[k] = v.Value()
	}

	return out, nil
}

// GetProperty returns a single hostname property
func (c *Conn) GetProperty(propertyName string) (interface{}, error) {
	var err error
	var prop dbus.Variant

	err = c.object.Call("org.freedesktop.DBus.Properties.Get", 0, dbusInterface, propertyName).Store(&prop)
	if err != nil {
		return nil, err
	}

	return prop.Value(), nil
}

// find the dynamic hostname
func (c *Conn) GetHostname() (string, error) {
	if hn, err := c.GetProperty("Hostname"); err != nil {
		return "", err
	} else if hns, ok := hn.(string); !ok {
		return "", fmt.Errorf("hostname has incorrect type: %T", hn)
	} else {
		return hns, nil
	}
}

// find the static hostname
func (c *Conn) GetStaticHostname() (string, error) {
	if hn, err := c.GetProperty("StaticHostname"); err != nil {
		return "", err
	} else if hns, ok := hn.(string); !ok {
		return "", fmt.Errorf("hostname has incorrect type: %T", hn)
	} else {
		return hns, nil
	}
}

// SetHostname asks hostnamed to set the hostname.
func (c *Conn) SetHostname(name string, askForAuth bool) error {
	return c.object.Call(dbusInterface+".SetHostname", 0, name, askForAuth).Err
}

// SetStaticHostname asks hostnamed to set the static hostname.
func (c *Conn) SetStaticHostname(name string, askForAuth bool) error {
	return c.object.Call(dbusInterface+".SetStaticHostname", 0, name, askForAuth).Err
}

// SetPrettyHostname asks hostnamed to set the pretty hostname.
func (c *Conn) SetPrettyHostname(name string, askForAuth bool) error {
	return c.object.Call(dbusInterface+".SetPrettyHostname", 0, name, askForAuth).Err
}

// SetIconName asks hostnamed to set the icon name following the XDG icon naming spec.
func (c *Conn) SetIconName(name string, askForAuth bool) error {
	return c.object.Call(dbusInterface+".SetIconName", 0, name, askForAuth).Err
}

// SetChassis asks hostnamed to set the chassis name.
func (c *Conn) SetChassis(name string, askForAuth bool) error {
	return c.object.Call(dbusInterface+".SetChassis", 0, name, askForAuth).Err
}
