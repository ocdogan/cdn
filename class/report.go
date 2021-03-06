package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

const version = "v3"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func sum(vals []interface{}) interface{} {
	switch vals[0].(type) {
	case int:
		var acc int
		for _, val := range vals {
			acc += val.(int)
		}
		return acc
	case float64:
		var acc float64
		for _, val := range vals {
			acc += val.(float64)
		}
		return acc
	default:
		fmt.Printf("??? - %T\n", vals[0])
		panic("unknown type")
	}
}

func avg(vals []interface{}) interface{} {
	res := sum(vals)
	switch res.(type) {
	case int:
		return res.(int) / len(vals)
	case float64:
		return res.(float64) / float64(len(vals))
	default:
		fmt.Printf("???? - %T\n", res)
		panic("unknown tyhpe")
	}
}

func main() {
	fh, err := os.Open("class/" + version + "/data.json")
	check(err)
	decoder := json.NewDecoder(fh)

	reports := make(map[string][]string)

	for err == nil {
		var obj map[string]interface{}
		err = decoder.Decode(&obj)
		if err != nil {
			break
		}

		// Lose the unique server prefixes
		lists := make(map[string][]interface{}, len(obj))
		for key, value := range obj {
			parts := strings.Split(key, ".")
			if len(parts) > 2 {
				parts = append([]string{parts[0]}, parts[2:]...)
			}
			key = strings.Join(parts, ".")
			lists[key] = append(lists[key], value)
		}

		// aggregate data as desired
		obj = make(map[string]interface{}, len(lists))
		for key, values := range lists {
			last := key[strings.LastIndex(key, ".")+1:]
			last = last[strings.LastIndex(last, "-")+1:]

			// summations
			switch last {
			case "bad":
				fallthrough
			case "img":
				fallthrough
			case "count":
				fallthrough
			case "cacheSize":
				fallthrough
			case "neighbor_list":
				fallthrough
			case "s2s_calls":
				fallthrough
			case "neighbor_miss":
				fallthrough
			case "neighbor_hit":
				fallthrough
			case "force_push":
				obj[key] = sum(values)

				// Average
			case "minute":
				fallthrough
			case "percentile":
				fallthrough
			case "rate":
				fallthrough
			case "max":
				fallthrough
			case "min":
				fallthrough
			case "dev":
				fallthrough
			case "mean":
				obj[key] = avg(values)
				if key == "client.render.max" {
					obj["count"] = len(values)
				}

				// identity
			case "uptime":
				obj[key] = values[0]

			default:
				fmt.Println("Unknown kehy", last)
			}
		}

		// TODO: generate a CSV for plotting
		reports["hit_miss"] = append(reports["hit_miss"], fmt.Sprintf(
			"%d, %f, %f", obj["count"], obj["server.neighbor_hit"], obj["server.neighbor_miss"],
		))
		reports["render"] = append(reports["render"], fmt.Sprintf(
			"%d, %f, %f, %f, %f, %f, %f, %f, %f", obj["count"], obj["client.render.min"], obj["client.render.mean"],
			obj["client.render.50-percentile"], obj["client.render.75-percentile"],
			obj["client.render.95-percentile"], obj["client.render.99-percentile"],
			obj["client.render.999-percentile"], obj["client.render.max"],
		))
		reports["client-server"] = append(reports["client-server"], fmt.Sprintf(
			"%d, %f, %f, %f, %f, %f", obj["count"], obj["client.request.count"],
			obj["origin.page.count"], obj["server.neighbor_hit"],
			obj["server.neighbor_miss"], obj["server.s2s_calls"],
		))
		fmt.Println("count", obj["count"])
	}

	for name, data := range reports {
		bits := strings.Join(data, "\n")
		err = ioutil.WriteFile("class/"+version+"/"+name+".csv", []byte(bits), 0644)
		check(err)
	}
}
