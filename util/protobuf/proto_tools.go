package protobuf

import (
	"bytes"
	"github.com/gogf/gf/v2/text/gstr"
	"io/ioutil"
	"os/exec"
	"path"
)

func IsProtocVersionOK() (bool, error) {
	var (
		err     error
		bytes   []byte
		version string
	)
	cmd := exec.Command("protoc", "--version")
	bytes, err = cmd.Output()
	if err != nil { //获取输出对象，可以从该对象中读取输出结果
		return false, err
	}
	version = string(bytes)
	return gstr.ContainsI(version, "protoc 3."), nil
}

func IsProtocTripleVersionOK() (bool, error) {
	var (
		err     error
		bytes   []byte
		version string
	)
	cmd := exec.Command("protoc-gen-go-triple", "--version")
	bytes, err = cmd.Output()
	if err != nil { //获取输出对象，可以从该对象中读取输出结果
		return false, err
	}
	version = string(bytes)
	return gstr.ContainsI(version, "protoc-gen-go-triple 1."), nil
}

func CallProtoc(curDir string, packageName string, goFileName string, seperatedPackage bool) error {
	var (
		protoPath string
		pbGoPath  string
		err       error
	)
	packageName = gstr.SubStrFromEx(packageName, "devops.gitlab.zfkunyu.com/cartsee-go/cartx-etl/")
	if seperatedPackage {
		protoPath = path.Join(curDir, packageName, goFileName, "proto")
	} else {
		protoPath = path.Join(curDir, packageName, "proto")
	}
	cmd := exec.Command("protoc", "-I"+protoPath, "--go_out=.", "--go-triple_out=.", goFileName+".proto")
	cmd.Dir = curDir
	_, err = cmd.Output()
	if err != nil {
		return err
	}
	if seperatedPackage {
		pbGoPath = path.Join(curDir, packageName, goFileName, "model", goFileName+".pb.go")
	} else {
		pbGoPath = path.Join(curDir, packageName, "model", goFileName+".pb.go")
	}
	err = ReplaceString(pbGoPath, ",omitempty", "")
	if err != nil {
		return err
	}
	return nil
}

func ReplaceString(fileName string, original, replacement string) error {
	input, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	output := bytes.Replace(input, []byte(original), []byte(replacement), -1)
	if err = ioutil.WriteFile(fileName, output, 0644); err != nil {
		return err
	}
	return nil
}
