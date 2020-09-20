// package main

// import (
// 	"archive/zip"
// 	"encoding/json"
// 	"flag"
// 	"fmt"
// 	"io/ioutil"
// 	"strings"
// )

// func main() {
// 	word := flag.String("word", "", "a word")
// 	flag.Parse()

// 	r, err := zip.OpenReader("foc-slack-export.zip")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer r.Close()

// 	for _, f := range r.File {
// 		if !f.FileInfo().IsDir() {
// 			// log.Println("------------------------>", f.FileInfo().IsDir(), f.Name, strings.Index(f.Name, ".json"), len(f.Name)-5)
// 			if strings.Index(f.Name, ".json") == len(f.Name)-5 {
// 				rc, err := f.Open()
// 				if err != nil {
// 					panic(err)
// 				}

// 				b, err := ioutil.ReadAll(rc)
// 				if err != nil {
// 					panic(err)
// 				}
// 				rc.Close()

// 				var all []map[string]interface{}
// 				err = json.Unmarshal(b, &all)
// 				if err != nil {
// 					panic(err)
// 				}

// 				for _, m := range all {

// 					if text, ok := m["text"].(string); ok {
// 						// log.Println("---------------->", text)
// 						words := strings.Split(text, " ")

// 						for i, w := range words {
// 							if w == *word {
// 								before := make([]string, 4)
// 								if i-4 >= 0 {
// 									before[0], before[1], before[2], before[3] = words[i-4], words[i-3], words[i-2], words[i-1]
// 								}

// 								after := make([]string, 4)
// 								if i+4 < len(words) {
// 									after[0], after[1], after[2], after[3] = words[i+1], words[i+2], words[i+3], words[i+4]
// 								}

// 								fmt.Printf("File %s : %s %s %s %s %s %s %s %s %s\n", f.Name, before[0], before[1], before[2], before[3], w, after[0], after[1], after[2], after[3])
// 							}
// 						}
// 					}
// 				}
// 			}
// 		}
// 	}
// }

package main

import (
	"archive/zip"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
)

var (
	g_key    string
	g_word   string
	g_length int
)

func main() {
	// The json key ("text") whose value we are searching to contain "-word". We remove the hardcoded (but default "text") and generalize it to any json key.
	key := flag.String("key", "text", "json key whose value is to be searched")

	// word is the value that we are searching for the json key.
	word := flag.String("word", "", "a word")

	// Number of words that you want before and after "-word" wherever found. We remove hardcoded (but default "4") and generalize it to take non negative numbers.
	length := flag.Uint("length", 4, "number of words neighbouring to -word")

	flag.Parse()

	// The length of word to search ("-word" input value) should be greater than zero.
	if word != nil && len(*word) == 0 {
		fmt.Println("Argument \"-word\" is mandatory")
		fmt.Println()
		flag.Usage()
		return
	}

	g_key = *key
	g_word = *word
	g_length = (int)(*length)

	r, err := zip.OpenReader("foc-slack-export.zip")
	if err != nil {
		panic(err)
	}
	defer r.Close()

	handle_file(r)
}

func handle_file(r *zip.ReadCloser) {

	for _, f := range r.File {
		if !f.FileInfo().IsDir() {
			if strings.Index(f.Name, ".json") == len(f.Name)-5 {
				rc, err := f.Open()
				if err != nil {
					panic(err)
				}

				b, err := ioutil.ReadAll(rc)
				if err != nil {
					panic(err)
				}
				rc.Close()

				var all []map[string]interface{}
				err = json.Unmarshal(b, &all)
				if err != nil {
					panic(err)
				}

				for _, m := range all {
					for m_key, m_val := range m {
						handle_json(f.Name, m_key, m_val)
					}
				}
			}
		}
	}
}

func handle_json(file_name, key string, inter interface{}) {

	switch inter.(type) {
	case string:
		if key == g_key {
			scan_words(file_name, inter.(string))
		}

	case map[string]interface{}:
		// log.Println("map[string]interface{} ##############", file_name, key, inter)
		m := inter.(map[string]interface{})
		for m_key, m_val := range m {
			handle_json(file_name, m_key, m_val)
		}

	case []map[string]interface{}:
		// log.Println("[]map[string]interface{} ##############", file_name, key, inter)
		s := inter.([]map[string]interface{})
		for _, m := range s {
			for m_key, m_val := range m {
				handle_json(file_name, m_key, m_val)
			}
		}

	case []interface{}:
		// log.Println("[]interface{} ##############", file_name, key, inter)
		inter_slice := inter.([]interface{})
		string_slice := make([]string, 0)

		for i := 0; i < len(inter_slice); i++ {
			switch inter_slice[i].(type) {
			// If the value if key ("-key") is a json string array then we concatenate the array into single string and then search given "-word" in that string.
			case string:
				string_slice = append(string_slice, inter_slice[i].(string))
			default:
				handle_json(file_name, key, inter_slice[i])
			}
		}

		if len(string_slice) > 0 {
			handle_json(file_name, key, strings.Join(string_slice, " "))
		}
	}
}

func scan_words(file_name, text string) {
	// Use strings.Fields() instead of strings.Split() this helps us to also avoid new line '\n'
	// and double spaces "  " and few other charecters too, apart from simple space " ".
	words := strings.Fields(text)

	for i, w := range words {
		// Consider using levenshtein edit distance ("github.com/agnivade/levenshtein") to include spell checks.
		// Removing comma at end of word if it occurs, more generic way should be to use levenshtein edit distance.
		// For now lets atleast make it case-insensitive with EqualFold()
		if strings.EqualFold(strings.TrimRight(w, ","), g_word) {
			stmt := make([]string, 0)

			index := i - g_length
			if index < 0 {
				index = 0
			}

			end := i + g_length

			for ; index <= end && index < len(words); index++ {
				stmt = append(stmt, words[index])
			}

			fmt.Printf("File %s : %s \n", file_name, strings.Join(stmt, " "))
		}
	}
}
