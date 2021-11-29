package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/blushft/go-diagrams/diagram"
	"github.com/blushft/go-diagrams/nodes/aws"
)

type datStruct struct {
	Label    string
	SubID    string
	SubName  string
	SGs      [][]string
	Filename string
}

type confStruct struct {
	Label string
	Type  string
	SG    string
	Name  string
}

const (
	toolName = "godia"
)

var (
	debug, logging, verbose bool
	path                    string
)

func main() {
	_ini := flag.String("ini", "godia.ini", "[-ini=configuration file (.ini) filename]")
	_Debug := flag.Bool("debug", false, "[-debug=debug mode (true is enable)]")
	_Dir := flag.String("dir", "dat", "[-dir=search directory]")
	_Output := flag.String("output", "nw", "[-output=output .dot filename]")
	_Verbose := flag.Bool("verbise", true, "[-verbose=incude id verbose (true is enable)]")
	_VPC := flag.String("vpc,", "vpc,vpc-00000000000000000", "[-vpc=vpc name and id (for Label)]")
	flag.Parse()
	debug = bool(*_Debug)
	verbose = bool(*_Verbose)

	switch runtime.GOOS {
	case "linux":
		path = "/"
	case "windows":
		path = "\\"
	}

	datfiles := dirwalk(*_Dir)
	if debug == true {
		fmt.Println(datfiles)
	}

	conf, err := iniRead(*_ini)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dats, err := readDat(datfiles)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		if debug == true {
			fmt.Println(dats)
		}

		os.RemoveAll("go-diagrams")
		drawDiagram(dats, *_Output, *_VPC, conf)
	}
	os.Exit(0)
}

func drawDiagram(dats []datStruct, dFilename string, VPC string, inis []confStruct) {
	var VPCt string
	vpcs := strings.Split(VPC, ",")
	if verbose == true {
		VPCt = vpcs[0] + "[" + vpcs[1] + "]"
	} else {
		VPCt = vpcs[0]
	}

	d, err := diagram.New(diagram.Filename(dFilename), diagram.Direction("LR"), diagram.Label(VPCt))
	if err != nil {
		log.Fatal(err)
	}

	nodes := make([][]*diagram.Node, len(dats))
	datas := make([]*diagram.Group, len(dats))

	if len(dats) == 1 {
		nodes[0] = make([]*diagram.Node, len(dats[0].SGs))

		for r, node := range dats[0].SGs {
			var tag *diagram.Node
			debugLog(node[0] + " " + node[1])
			if verbose == true {
				tag = aws.Compute.Ec2(diagram.NodeLabel(node[0] + "[" + node[1] + "]"))
			} else {
				tag = aws.Compute.Ec2(diagram.NodeLabel(node[0]))
			}
			nodes[0][r] = tag
		}

		datas[0] = diagram.NewGroup(dats[0].Label)
		if verbose == true {
			datas[0].NewGroup(dats[0].Label).
				Label(dats[0].Label + ": " + dats[0].SubName + "[" + dats[0].SubID + "]").Add(nodes[0]...)
		} else {
			datas[0].NewGroup(dats[0].Label).
				Label(dats[0].Label + ": " + dats[0].SubName).Add(nodes[0]...)
		}

		d.Group(datas[0])
	} else {
		for i, dat := range dats {
			sameFlag := false
			nodes[i] = make([]*diagram.Node, len(dats[i].SGs))

			for r, node := range dat.SGs {
				strVal := strconv.Itoa(i)
				var tag *diagram.Node
				debugLog(strVal + " " + node[0] + " " + node[1])
				if verbose == true {
					tag = aws.Compute.Ec2(diagram.NodeLabel(node[0] + "[" + node[1] + "]"))
				} else {
					tag = aws.Compute.Ec2(diagram.NodeLabel(node[0]))
				}
				nodes[i][r] = tag
			}

			if sameFlag == false {
				datas[i] = diagram.NewGroup(dats[i].Label)
				if verbose == true {
					datas[i].NewGroup(dats[i].Label).
						Label(dats[i].Label + ": " + dats[i].SubName + "[" + dats[i].SubID + "]").
						Add(nodes[i]...)
				} else {
					datas[i].NewGroup(dats[i].Label).
						Label(dats[i].Label + ": " + dats[i].SubName).
						Add(nodes[i]...)
				}
				sameFlag = true
			} else {
				if verbose == true {
					datas[i].
						Label(dats[i].Label + ": " + dats[i].SubName + "[" + dats[i].SubID + "]").
						Add(nodes[i]...)
				} else {
					datas[i].
						Label(dats[i].Label + ": " + dats[i].SubName).
						Add(nodes[i]...)
				}
			}
		}

		for i := 0; i < len(nodes)-1; i++ {
			for _, node1 := range nodes[i] {
				for _, node2 := range nodes[i+1] {
					if node1.Options.Label == node2.Options.Label {
						d.Connect(node1, node2, diagram.Forward()).Group(datas[i+1])
					}
				}
			}
			d.Group(datas[i])
		}
	}

	inets := make([]*diagram.Node, len(inis))
	for x, ini := range inis {
		inets[x] = aws.Network.InternetGateway(diagram.NodeLabel(ini.Name))
		for y, dat := range dats {
			for r, sg := range dat.SGs {
				if ini.Label == dat.Filename && ini.SG == sg[0] {
					if ini.Type == "I" {
						d.Connect(inets[x], nodes[y][r], diagram.Forward()).Group(datas[y])
					} else if ini.Type == "O" {
						d.Connect(inets[x], nodes[y][r], diagram.Reverse()).Group(datas[y])
					} else if ini.Type == "D" {
						d.Connect(inets[x], nodes[y][r], diagram.Bidirectional()).Group(datas[y])
					}
				}
			}
		}
	}

	if err := d.Render(); err != nil {
		log.Fatal(err)
	}
}

func iniRead(inifile string) ([]confStruct, error) {
	var ini []confStruct

	debugLog("ini file: " + inifile)

	fp, err := os.Open(inifile)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	scanner := bufio.NewScanner(fp)

	for scanner.Scan() {
		stra := scanner.Text()
		if len(stra) > 0 {
			strb := strings.Split(stra, ",")
			ini = append(ini, confStruct{Label: strb[0], Type: strb[1], SG: strb[2], Name: strb[3]})
		}
	}

	if err = scanner.Err(); err != nil {
		return nil, err
	}

	return ini, nil
}

func readDat(datfiles []string) ([]datStruct, error) {
	var dats []datStruct

	for _, file := range datfiles {
		debugLog("file: " + file)

		fp, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer fp.Close()

		var Label string
		Labels := strings.Split(file, path)
		if strings.Index(Labels[len(Labels)-1], "_") != -1 {
			Label = strings.Split(Labels[len(Labels)-1], "_")[1]
		} else {
			Label = Labels[len(Labels)-1]
		}

		reader := bufio.NewReaderSize(fp, 4096)
		lineA, err := reader.ReadString('\n')
		lineA = strings.TrimRight(lineA, "\n")
		lineA = strings.TrimRight(lineA, "\r")
		if err != nil {
			return nil, err
		}
		stra := strings.Split(lineA, ",")

		lineB, err := reader.ReadString('\n')
		lineB = strings.TrimRight(lineB, "\n")
		lineB = strings.TrimRight(lineB, "\r")
		if err != nil {
			return nil, err
		}
		strb := strings.Split(lineB, " ")
		var stre [][]string
		for _, dst := range strb {
			var strd []string
			strc := strings.Split(dst, ",")
			strd = append(strd, strc[0])
			strd = append(strd, strc[1])
			stre = append(stre, strd)
		}

		dats = append(dats, datStruct{Label: Label, SubName: stra[0], SubID: stra[1], SGs: stre, Filename: Labels[len(Labels)-1]})
	}
	return dats, nil
}

func dirwalk(dir string) []string {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}

	var paths []string
	for _, file := range files {
		if file.IsDir() {
			paths = append(paths, dirwalk(filepath.Join(dir, file.Name()))...)
			continue
		}
		paths = append(paths, filepath.Join(dir, file.Name()))
	}

	return paths
}

func debugLog(message string) {
	var file *os.File
	var err error

	if debug == true {
		fmt.Println(message)
	}

	if logging == false {
		return
	}

	const layout = "2006-01-02_15"
	t := time.Now()
	filename := toolName + "_" + t.Format(layout) + ".log"

	if Exists(filename) == true {
		file, err = os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0666)
	} else {
		file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	}

	if err != nil {
		log.Fatal(err)
		return
	}
	defer file.Close()
	fmt.Fprintln(file, message)
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
