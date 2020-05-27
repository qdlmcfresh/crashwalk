package main

import (
	"flag"
	"fmt"
	"github.com/bnagy/crashwalk/crash"
	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
	"log"
	"os"
	"path"
	"encoding/json"
	"os/exec"
)

type summary struct {
	detail string
	count  int
}

type crashJsonOutput struct {
	Count int
	CrashEntry crash.Crash
	Uname string
}

func uname() string{
	app := "/bin/uname"
	arg0 := "-a"
	cmd := exec.Command(app,arg0)
	stdout, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(stdout)

}

func main() {

	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"\n  %s summarizes crashes in the given crashwalk databases by major.minor hash\n",
			path.Base(os.Args[0]),
		)
		fmt.Fprintf(
			os.Stderr,
			"  Usage: %s /path/to/crashwalk.db [db db ...]\n\n",
			path.Base(os.Args[0]),
		)
	}

	flag.Parse()
	for _, arg := range flag.Args() {

		db, err := bolt.Open(arg, 0600, nil)
		if err != nil {
			log.Fatalf("failed to open DB (%s): %s", arg, err)
		}
		defer db.Close()

		db.View(func(tx *bolt.Tx) error {

			tx.ForEach(func(name []byte, b *bolt.Bucket) error {
				crashes := make(map[string]crashJsonOutput)
				b.ForEach(func(k, v []byte) error {
					ce := &crash.Entry{}
					err := proto.Unmarshal(v, ce)
					if err != nil {
						return err
					}
					c, ok := crashes[ce.Hash]
					if !ok {
						//Start with 0 because we always increment
						c = crashJsonOutput{Count: 0, CrashEntry: crash.Crash{Entry: *ce}, Uname: uname()}
					}
					c.Count++
					crashes[ce.Hash] = c
					return nil
				})

				j, err := json.MarshalIndent(crashes, "", "\t")
				if err != nil {
					fmt.Println("Error while creating json output")
					return nil
				}
				fmt.Println(string(j))
				return nil
			})
			return nil
		})
	}
}
