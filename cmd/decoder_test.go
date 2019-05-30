package cmd

import (
	. "github.com/pingcap/check"
)

var _ = Suite(&decoderTestSuite{})

type decoderTestSuite struct{}

func (s *decoderTestSuite) TestTableRowDecode(c *C) {
	cmd := initCommand()
	args := []string{"decoder", "t\x80\x00\x00\x00\x00\x00\x07\x8f_r\x80\x00\x00\x00\x00\x08\x3b\xba"}
	_, output, err := executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Check(string(output), Equals, "format: table_row\ntable_id: 1935\nrow_id: 539578\n")

	args = []string{"decoder", "t\200\000\000\000\000\000\007\217_r\200\000\000\000\000\010;\272"}
	_, output, err = executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Check(string(output), Equals, "format: table_row\ntable_id: 1935\nrow_id: 539578\n")
}

func (s *decoderTestSuite) TestTableIndexDecode(c *C) {
	cmd := initCommand()
	args := []string{"decoder", "t\x80\x00\x00\x00\x00\x00\x00\x5f_i\x80\x00\x00\x00\x00\x00\x00\x01\x03\x80\x00\x00\x00\x00\x00\x00\x02\x03\x80\x00\x00\x00\x00\x00\x00\x02"}
	_, output, err := executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Check(string(output), Equals, "format: table_index\n"+
		"table_id: 95\n"+
		"row_id: 1\n"+
		"index_value[0]: {type: bigint, value: 2}\n"+
		"index_value[1]: {type: bigint, value: 2}\n")
}

func (s *decoderTestSuite) TestIndexValueDecode(c *C) {
	cmd := initCommand()
	args := []string{"decoder", "CAQCBmFiYw=="}
	_, output, err := executeCommandC(cmd, args...)
	c.Assert(err, IsNil)
	c.Check(string(output), Equals, "format: index_value\n"+
		"index_value[0]: {type: bigint, value: 2}\n"+
		"index_value[1]: {type: bytes, value: abc}\n")
}