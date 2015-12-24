// Copyright 2015 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"encoding/json"
	"flag"
	"testing"
	"time"

	"github.com/ngaut/go-zookeeper/zk"
	"github.com/ngaut/zkhelper"
	. "github.com/pingcap/check"
)

var testZKAddr = flag.String("zk", "127.0.0.1:2181", "test zookeeper address")

func TestUtil(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&testUtilSuite{})

type testUtilSuite struct {
	zkConn zkhelper.Conn

	rootPath string
}

func (s *testUtilSuite) SetUpSuite(c *C) {
	conn, err := zkhelper.ConnectToZkWithTimeout(*testZKAddr, time.Second)
	c.Assert(err, IsNil)
	s.zkConn = conn

	s.rootPath = "/zk/tso_util_test"

}

func (s *testUtilSuite) TearDownSuite(c *C) {
	if s.zkConn != nil {
		err := zkhelper.DeleteRecursive(s.zkConn, s.rootPath, -1)
		c.Assert(err, IsNil)

		s.zkConn.Close()
	}
}

func (s *testUtilSuite) TestLeader(c *C) {
	conn, err := zkhelper.ConnectToZkWithTimeout(*testZKAddr, time.Second)
	c.Assert(err, IsNil)
	defer conn.Close()

	leaderPath := getLeaderPath(s.rootPath)
	conn.Delete(leaderPath, -1)

	_, err = GetLeader(conn, s.rootPath)
	c.Assert(err, NotNil)

	_, _, err = GetWatchLeader(conn, s.rootPath)
	c.Assert(err, NotNil)

	_, err = zkhelper.CreateRecursive(conn, leaderPath, "", 0, zk.WorldACL(zkhelper.PERM_FILE))
	c.Assert(err, IsNil)

	_, err = GetLeader(conn, s.rootPath)
	c.Assert(err, NotNil)

	_, _, err = GetWatchLeader(conn, s.rootPath)
	c.Assert(err, NotNil)

	addr := "127.0.0.1:1234"
	m := map[string]interface{}{
		"Addr": addr,
	}

	data, err := json.Marshal(m)
	c.Assert(err, IsNil)

	_, err = conn.Set(leaderPath, data, -1)
	c.Assert(err, IsNil)

	v, err := GetLeader(conn, s.rootPath)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, addr)

	v, _, err = GetWatchLeader(conn, s.rootPath)
	c.Assert(err, IsNil)
	c.Assert(v, Equals, addr)
}
