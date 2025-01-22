package main

import (
	"archive/tar"
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	buffered := bufio.NewReader(os.Stdin)

	headers := []*tar.Header{}

	{
		reader := tar.NewReader(buffered)

		for {
			header, err := reader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}

			headers = append(headers, header)
		}
	}

	suppress := map[string]bool{}
	transfer := map[string][]func(*tar.Header){}
	trailing := map[string][]func(*tar.Writer){}

	{
		append_suppress := (func(key string) {
			suppress[key] = true
		})

		append_transfer := (func(key string, fun func(*tar.Header)) {
			transfer[key] = append(transfer[key], fun)
		})

		append_trailing := (func(key string, fun func(*tar.Writer)) {
			trailing[key] = append(trailing[key], fun)
		})

		for _, header := range headers {
			path := header.Name
			base := filepath.Base(path)
			dirn := filepath.Dir(path) + "/"

			if base == ".wh..wh..opq" {
				append_suppress(path)

				append_transfer(dirn, (func(header *tar.Header) {
					if header.PAXRecords == nil {
						header.PAXRecords = map[string]string{}
					}

					header.PAXRecords["SCHILY.xattr.trusted.overlay.opaque"] = "y"
				}))
			} else if strings.HasPrefix(base, ".wh.") {
				originalBase := base[len(".wh."):]
				originalPath := filepath.Join(dirn, originalBase)

				append_suppress(path)

				append_trailing(dirn, (func(writer *tar.Writer) {
					err := writer.WriteHeader(&tar.Header{
						Typeflag:   tar.TypeChar,
						Name:       originalPath,
						Mode:       header.Mode,
						Uid:        header.Uid,
						Gid:        header.Gid,
						Uname:      header.Uname,
						Gname:      header.Gname,
						ModTime:    header.ModTime,
						AccessTime: header.AccessTime,
						ChangeTime: header.ChangeTime,
						Format:     tar.FormatPAX,
					})
					if err != nil {
						panic(err)
					}
				}))
			}
		}
	}

	// skip zeroes between archives
	{
		for {
			val, err := buffered.ReadByte()
			if err != nil {
				panic(nil)
			}
			if val != 0 {
				buffered.UnreadByte()
				break
			}
		}
	}

	{
		reader := tar.NewReader(buffered)
		writer := tar.NewWriter(os.Stdout)

		for {
			header, err := reader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}

			path := header.Name

			if suppress[path] {
				continue
			}

			header.Format = tar.FormatPAX

			for _, fun := range transfer[path] {
				fun(header)
			}

			writer.WriteHeader(header)

			_, err := io.Copy(writer, reader)
			if err != nil {
				panic(err)
			}

			for _, fun := range trailing[path] {
				fun(writer)
			}
		}
	}
}
