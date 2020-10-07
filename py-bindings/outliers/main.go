package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/christian-korneck/go-python3"
)

//detect wraps around the gen_testdata() function from the given Python module. Borrows the PyObject reference.
func detect(module *python3.PyObject, data []float64) ([]int, error) {

	pylist := python3.PyList_New(len(data)) //retval: New reference, gets stolen later
	for i := 0; i < len(data); i++ {
		item := python3.PyFloat_FromDouble(data[i]) //retval: New reference, gets stolen later
		ret := python3.PyList_SetItem(pylist, i, item)
		if ret != 0 {
			if python3.PyErr_Occurred() != nil {
				python3.PyErr_Print()
			}
			item.DecRef()
			pylist.DecRef()
			return nil, fmt.Errorf("error setting list item")
		}
	}

	args := python3.PyTuple_New(1) //retval: New reference
	if args == nil {
		pylist.DecRef()
		return nil, fmt.Errorf("error creating args tuple")
	}
	defer args.DecRef()
	ret := python3.PyTuple_SetItem(args, 0, pylist) //steals ref to pylist
	pylist = nil
	if ret != 0 {
		if python3.PyErr_Occurred() != nil {
			python3.PyErr_Print()
		}
		pylist.DecRef()
		pylist = nil
		return nil, fmt.Errorf("error setting args tuple item")
	}

	oDict := python3.PyModule_GetDict(module) //retval: Borrowed
	if !(oDict != nil && python3.PyErr_Occurred() == nil) {
		python3.PyErr_Print()
		return nil, fmt.Errorf("could not get dict for module")
	}
	detect := python3.PyDict_GetItemString(oDict, "detect")
	if !(detect != nil && python3.PyCallable_Check(detect)) { //retval: Borrowed
		return nil, fmt.Errorf("could not find function 'detect'")
	}
	detectdataPy := detect.CallObject(args)
	if !(detectdataPy != nil && python3.PyErr_Occurred() == nil) { //retval: New reference
		python3.PyErr_Print()
		return nil, fmt.Errorf("error calling function detect")
	}
	defer detectdataPy.DecRef()
	outliers, err := goSliceFromPylist(detectdataPy, "int", false)
	if err != nil {
		return nil, fmt.Errorf("error converting pylist to go list: %s", err)
	}

	return outliers.([]int), nil

}

//goSliceFromPylist converts a []float64 pylist to a go list. Borrows the PyObject reference.
func goSliceFromPylist(pylist *python3.PyObject, itemtype string, strictfail bool) (interface{}, error) {

	seq := pylist.GetIter() //ret val: New reference
	if !(seq != nil && python3.PyErr_Occurred() == nil) {
		python3.PyErr_Print()
		return nil, fmt.Errorf("error creating iterator for list")
	}
	defer seq.DecRef()
	tNext := seq.GetAttrString("__next__") //ret val: new ref
	if !(tNext != nil && python3.PyCallable_Check(tNext)) {
		return nil, fmt.Errorf("iterator has no __next__ function")
	}
	defer tNext.DecRef()

	var golist interface{}
	var compare *python3.PyObject
	switch itemtype {
	case "float64":
		golist = []float64{}
		compare = python3.PyFloat_FromDouble(0)
	case "int":
		golist = []int{}
		compare = python3.PyLong_FromGoInt(0)
	}
	if compare == nil {
		return nil, fmt.Errorf("error creating compare var")
	}
	defer compare.DecRef()

	pytype := compare.Type() //ret val: new ref
	if pytype == nil && python3.PyErr_Occurred() != nil {
		python3.PyErr_Print()
		return nil, fmt.Errorf("error getting type of compare var")
	}
	defer pytype.DecRef()

	errcnt := 0

	pylistLen := pylist.Length()
	if pylistLen == -1 {
		return nil, fmt.Errorf("error getting list length")
	}

	for i := 1; i <= pylistLen; i++ {
		item := tNext.CallObject(nil) //ret val: new ref
		if item == nil && python3.PyErr_Occurred() != nil {
			python3.PyErr_Print()
			return nil, fmt.Errorf("error getting next item in sequence")
		}
		itemType := item.Type()
		if itemType == nil && python3.PyErr_Occurred() != nil {
			python3.PyErr_Print()
			return nil, fmt.Errorf("error getting item type")
		}

		defer itemType.DecRef()

		if itemType != pytype {
			//item has wrong type, skip it
			if item != nil {
				item.DecRef()
			}
			errcnt++
			continue
		}

		switch itemtype {
		case "float64":
			itemGo := python3.PyFloat_AsDouble(item)
			if itemGo != -1 && python3.PyErr_Occurred() == nil {
				golist = append(golist.([]float64), itemGo)
			} else {
				if item != nil {
					item.DecRef()
				}
				errcnt++
			}
		case "int":
			itemGo := python3.PyLong_AsLong(item)
			if itemGo != -1 && python3.PyErr_Occurred() == nil {
				golist = append(golist.([]int), itemGo)
			} else {
				if item != nil {
					item.DecRef()
				}
				errcnt++
			}
		}

		if item != nil {
			item.DecRef()
			item = nil
		}
	}
	if errcnt > 0 {
		if strictfail {
			return nil, fmt.Errorf("could not add %d values (wrong type?)", errcnt)
		}
	}

	return golist, nil
}

//genTestdata wraps around the gen_testdata() function from the given Python module. Borrows the PyObject reference.
func genTestdata(module *python3.PyObject) ([]float64, error) {
	oDict := python3.PyModule_GetDict(module) //ret val: Borrowed
	if !(oDict != nil && python3.PyErr_Occurred() == nil) {
		python3.PyErr_Print()
		return nil, fmt.Errorf("could not get dict for module")
	}
	genTestdata := python3.PyDict_GetItemString(oDict, "gen_testdata") //retval: Borrowed
	if !(genTestdata != nil && python3.PyCallable_Check(genTestdata)) {
		return nil, fmt.Errorf("could not find function 'gen_testdata'")
	}
	testdataPy := genTestdata.CallObject(nil) //retval: New reference
	if !(testdataPy != nil && python3.PyErr_Occurred() == nil) {
		python3.PyErr_Print()
		return nil, fmt.Errorf("error calling function gen_testdata")
	}
	defer testdataPy.DecRef()
	testdataGo, err := goSliceFromPylist(testdataPy, "float64", false)
	if err != nil {
		return nil, fmt.Errorf("error converting pylist to go list: %s", err)
	}

	return testdataGo.([]float64), nil
}

func demo(module *python3.PyObject) {

	testdata, err := genTestdata(module)
	if err != nil {
		log.Fatalf("Error getting testdata: %s", err)
	}

	outliers, err := detect(module, testdata)
	if err != nil {
		log.Fatalf("Error detecting outliers: %s", err)
	}
	fmt.Println(outliers)

}

func main() {

	defer python3.Py_Finalize()
	python3.Py_Initialize()
	if !python3.Py_IsInitialized() {
		fmt.Println("Error initializing the python interpreter")
		os.Exit(1)
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	// we could also use PySys_GetObject("path") + PySys_SetPath,
	//but this is easier (at the cost of less flexible error handling)
	ret := python3.PyRun_SimpleString("import sys\nsys.path.append(\"" + dir + "\")")
	if ret != 0 {
		log.Fatalf("error appending '%s' to python sys.path", dir)
	}

	oImport := python3.PyImport_ImportModule("pyoutliers") //ret val: new ref
	if !(oImport != nil && python3.PyErr_Occurred() == nil) {
		python3.PyErr_Print()
		log.Fatal("failed to import module 'pyoutliers'")
	}

	defer oImport.DecRef()

	oModule := python3.PyImport_AddModule("pyoutliers") //ret val: borrowed ref (from oImport)

	if !(oModule != nil && python3.PyErr_Occurred() == nil) {
		python3.PyErr_Print()
		log.Fatal("failed to add module 'pyoutliers'")
	}

	demo(oModule)

}
