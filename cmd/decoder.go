// Copyright 2018 PingCAP, Inc.
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

package cmd

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/pingcap/tidb/types"
	"github.com/pingcap/tidb/util/codec"
	"github.com/spf13/cobra"
)

var (
	keyFormat string
	keyValue  string
)

const (
	formatFlagName = "format"
	keyFlagName    = "key"
)

// decoderCmd represents the key-decoder command
var decoderCmd = &cobra.Command{
	Use:   "decoder",
	Short: "decode key",
	Long: `decode --format (tabel_row/table_index/value) --key (key)
	table_row:   key format like 'txxx_rxxx'
	table_index: key format like 'txxx_ixxx'
	value:       base64 encoded value`,
	RunE: decodeKeyFunc,
}

func decodeKey(text string) (string, error) {
	var buf []byte
	r := bytes.NewBuffer([]byte(text))
	for {
		c, err := r.ReadByte()
		if err != nil {
			if err != io.EOF {
				return "", err
			}
			break
		}
		if c != '\\' {
			buf = append(buf, c)
			continue
		}
		n := r.Next(1)
		if len(n) == 0 {
			return "", io.EOF
		}
		// See: https://golang.org/ref/spec#Rune_literals
		if idx := strings.IndexByte(`abfnrtv\'"`, n[0]); idx != -1 {
			buf = append(buf, []byte("\a\b\f\n\r\t\v\\'\"")[idx])
			continue
		}

		switch n[0] {
		case 'x':
			fmt.Sscanf(string(r.Next(2)), "%02x", &c)
			buf = append(buf, c)
		default:
			n = append(n, r.Next(2)...)
			_, err := fmt.Sscanf(string(n), "%03o", &c)
			if err != nil {
				return "", err
			}
			buf = append(buf, c)
		}
	}
	return string(buf), nil
}

func decodeIndexValue(buf []byte) {
	key := buf
	for len(key) > 0 {
		remain, d, e := codec.DecodeOne(key)
		if e != nil {
			break
		} else {
			s, _ := d.ToString()
			typeStr := types.KindStr(d.Kind())
			fmt.Printf("type: %v, value: %v\n", typeStr, s)
		}
		key = remain
	}
}

func decodeTableIndex(buf []byte) error {
	if len(buf) >= 19 && buf[0] == 't' && buf[9] == '_' && buf[10] == 'i' {
		table_id := buf[1:9]
		row_id := buf[11:19]
		indexValue := buf[19:]
		_, tableID, _ := codec.DecodeInt(table_id)
		fmt.Printf("table_id: %v\n", tableID)
		_, rowID, _ := codec.DecodeInt(row_id)
		fmt.Printf("index_id: %v\n", rowID)
		decodeIndexValue(indexValue)
		return nil
	}
	return fmt.Errorf("illegal code format")
}

func decodeTableRow(buf []byte) error {
	if len(buf) == 19 && buf[0] == 't' && buf[9] == '_' && buf[10] == 'r' {
		table_id := buf[1:9]
		row_id := buf[11:]
		_, tableID, _ := codec.DecodeInt(table_id)
		fmt.Printf("table_id: %v\n", tableID)
		_, rowID, _ := codec.DecodeInt(row_id)
		fmt.Printf("row_id: %v\n", rowID)
		return nil
	}
	return fmt.Errorf("illegal code format")
}

func decodeKeyFunc(_ *cobra.Command, args []string) error {
	if len(args) > 2 {
		return fmt.Errorf("too many arguments")
	}
	if keyFormat == "" {
		return fmt.Errorf("format argument can not be null")
	}
	if keyValue == "" {
		return fmt.Errorf("no key to decode")
	}
	if keyFormat == "table_row" {
		raw, _ := decodeKey(keyValue)
		err := decodeTableRow([]byte(raw))
		return err
	} else if keyFormat == "table_index" {
		raw, _ := decodeKey(keyValue)
		err := decodeTableIndex([]byte(raw))
		return err
	} else if keyFormat == "value" {
		b64decode, err := base64.StdEncoding.DecodeString(keyValue)
		if err != nil {
			return err
		}
		decodeIndexValue(b64decode)
	}
	return nil
}

func init() {
	decoderCmd.Flags().StringVarP(&keyFormat, formatFlagName, "f", "", "the key format you want decode")
	decoderCmd.Flags().StringVarP(&keyValue, keyFlagName, "k", "", "the key you want decode")
}
