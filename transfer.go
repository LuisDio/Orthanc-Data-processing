package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"time"
)

type BasicAuthTransport struct {
	Username string
	Password string
}

func (bat BasicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s",
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",
			bat.Username, bat.Password)))))
	return http.DefaultTransport.RoundTrip(req)
}

func (bat *BasicAuthTransport) Client() *http.Client {
	return &http.Client{Transport: bat}
}

func accessor() *http.Client {
	identification := BasicAuthTransport{"vizyon", "vercors"}
	connector := identification.Client()
	return connector

}

func contains(metadataList []string, metadataName string) (bool, string) {
	for _, element := range metadataList {
		if metadataName == element {
			return true, element
		}
	}
	return false, ""
}

func execDownload(hash string) {
	//downloadCm := "swarm --bzzapi http://localhost:301 down bzz:/" + hash + " data.zip"
	fmt.Println("Swarm Download execution")
	swarmCmd := exec.Command("bash", "-c", "./execDown.sh"+" "+hash)

	_, err := swarmCmd.Output()
	if err != nil {
		//panic(err)
		log.Fatal(err)
	}
	//fmt.Println(out1)
}

func execExtract() {
	fmt.Println("Os execute unzip file")
	unzipCmd := exec.Command("bash", "-c", "unzip -j /home/$USER/Download/data.zip -d /home/$USER/Download")

	_, err := unzipCmd.Output()
	if err != nil {
		panic(err)
	}
}

func execTransfer() {
	fmt.Println("Os execute Send file")
	transfCmd := exec.Command("bash", "-c", "storescu -aec ORTHANC localhost 4242 /home/$USER/Download/*.dcm")

	_, err := transfCmd.Output()
	if err != nil {
		panic(err)
	}
}

func execFileRemove() {
	fmt.Println("Os execute Remove dicom file")
	rmDicomCmd := exec.Command("bash", "-c", "rm /home/$USER/Download/*.dcm; rm /home/$USER/Download/*.zip")

	_, err := rmDicomCmd.Output()
	if err != nil {
		panic(err)
	}

}

func metadataAcess() {
	fmt.Println("Requesting....")
	db := accessor()
	reqst1, err := db.Get("http://192.168.0.174/orthanc/studies")

	if err != nil {
		log.Fatal(err)
	}
	responseData, err := ioutil.ReadAll(reqst1.Body)
	if err != nil {
		log.Fatal(err)
	}
	reqst1.Body.Close()

	fmt.Println("***************After converting request No 1**************")
	var study []string
	_ = json.Unmarshal([]byte(responseData), &study)

	//************************************
	// Type of Data from the newly converted request
	fmt.Printf("type of body data for request study is %T \n", study)
	fmt.Println("The number of studies is: ", len(study))

	//************************************

	for _, element := range study {
		fmt.Println(" ")
		fmt.Println("For this study id: ", element)
		time.Sleep(time.Second)

		reqst2, err := db.Get("http://192.168.0.174/orthanc/studies/" + element + "/metadata")
		if err != nil {
			log.Fatal(err)
		}
		respMetadata, err := ioutil.ReadAll(reqst2.Body)
		if err != nil {
			log.Fatal(err)
		}
		reqst2.Body.Close()

		fmt.Println("***************After converting request No 2**************")
		var metadata []string
		_ = json.Unmarshal([]byte(respMetadata), &metadata)

		// Type of Data from the newly converted request
		fmt.Printf("type of body data for request metadata is %T \n", metadata)
		fmt.Println("The number of metadata is: ", len(metadata))
		fmt.Println("The metadata list is ", metadata)

		// Desired Metadata where to look in
		iscontained, metaName := contains(metadata, "HHashValue") // Dont forget to change metadata to the HashValue

		if iscontained {
			// Implement a post for state when a data already processed
			fmt.Println("HPatientState found with name ", metaName)
			reqst3, err := db.Get("http://192.168.0.174/orthanc/studies/" + element + "/metadata/" + metaName)
			if err != nil {
				log.Fatal(err)
			}
			metaContent, err := ioutil.ReadAll(reqst3.Body)
			if err != nil {
				log.Fatal(err)
			}
			reqst3.Body.Close()
			fmt.Println("***************After converting request No 3**************")
			hash := string(metaContent)
			fmt.Println("The hash value is: ", hash)
			fmt.Println("Os call execute command")
			// 1-Exec_command to Download file from swarm to specific directory
			execDownload(hash)
			// 2-From chosen directory, unzip file in
			execExtract()
			// 3-Send All dicom file to local rohan
			execTransfer()
			// 5-Delete the previous file already sent
			execFileRemove()

		} else {
			//fmt.Println("PASS")
			continue
		}

	}

	//reqst_1.Body.Close()
}

func main() {
	metadataAcess()

}
