/*
 * Copyright 2020, 2021, 2022 Hewlett Packard Enterprise Development LP
 * Other additional copyright holders may be indicated within.
 *
 * The entirety of this work is licensed under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 *
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logging

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

type Parent struct {
	Name   string
	Value  string
	Logger *logrus.Entry
}

type Child struct {
	Name   string
	Parent *Parent
	Logger *logrus.Entry
}

func NewParent(logger *logrus.Logger) *Parent {
	p := &Parent{
		Name:  "Parent",
		Value: "Parent Value",
	}

	p.Logger = logger.WithFields(logrus.Fields{
		"Parent Name":  p.Name,
		"Parent Value": p.Value,
	})

	return p
}

func NewChild(parent *Parent, name string) *Child {
	c := &Child{
		Name:   name,
		Parent: parent,
	}

	c.Logger = parent.Logger.WithFields(logrus.Fields{
		"Child Name": c.Name,
	})

	return c
}

type CustomLoggingHook struct {
	Logger *logrus.Logger
}

func (c *CustomLoggingHook) Fire(entry *logrus.Entry) error {
	entry = entry.Dup()
	entry.Logger = c.Logger
	str, err := entry.String()
	fmt.Printf("Custom Logging Hook: %s", str)
	return err
}

func (*CustomLoggingHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
	}
}

func NewLoggingHook() logrus.Hook {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return &CustomLoggingHook{Logger: logger}
}

func TestLoggingStructure(t *testing.T) {
	log := logrus.New()
	log.SetLevel(logrus.TraceLevel)

	p := NewParent(log)
	c1 := NewChild(p, "child1")
	c2 := NewChild(p, "child2")

	p.Logger.Debug("Parent Log Entry")
	c1.Logger.Debug("Child Log Entry")
	time.Sleep(time.Second)
	c2.Logger.Debug("Child Log Entry")

	log.AddHook(NewLoggingHook())

	log.Debug("No Hook")
	log.Warn("No Hook")
	log.Error("Should Hook")
}
