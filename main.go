package main

import (

	"strings"
	"time"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"golang.org/x/oauth2/jwt"
)

func ServiceAccount(secretFile string) *http.Client {
	b, err := ioutil.ReadFile(secretFile)
	if err != nil {
		log.Fatal("error while reading the credential file", err)
	}
	var s = struct {
		Email      string `json:"client_email"`
		PrivateKey string `json:"private_key"`
	}{}
	json.Unmarshal(b, &s)
	config := &jwt.Config{
		Email:      s.Email,
		PrivateKey: []byte(s.PrivateKey),
		Scopes: []string{
			drive.DriveScope,
		},
		TokenURL: google.JWTTokenURL,
	}
	client := config.Client(context.Background())
	return client
}

func createFile(service *drive.Service, name string, mimeType string, content io.Reader, parentId string) (*drive.File, error) {
	f := &drive.File{
		MimeType: mimeType,
		Name:     name,
		Parents:  []string{parentId},
	}
	file, err := service.Files.Create(f).Media(content).Do()

	if err != nil {
		log.Println("Could not create file: " + err.Error())
		return nil, err
	}

	return file, nil
}

func main() {

	req, err := http.NewRequest("GET", "https://confluence.hflabs.ru/pages/viewpage.action?pageId=1181220999", nil)
	if err != nil {
		log.Fatal("Error reading request: ", err)
	}
	req.Header.Set("locale", "ru-RU")

	client := &http.Client{Timeout: time.Second * 1}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error reading response: ", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading body: ", err)
	}

	firstIndex := strings.Index(string(body), "<div class=\"table-wrap\">")
	lastIndex := strings.Index(string(body), "</p></td></tr></tbody></table></div>") + len("</p></td></tr></tbody></table></div>")
	result := string(body[firstIndex:lastIndex])
	fmt.Println(result)
	file, err := os.Create("data.xlsx")

	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer file.Close()
	file.WriteString(result)

	f, err := os.Open("data.xlsx")

	if err != nil {
		panic(fmt.Sprintf("cannot open file: %v", err))
	}

	defer f.Close()

	clientg := ServiceAccount("client-credentials.json")

	srv, err := drive.New(clientg)
	if err != nil {
		log.Fatalf("Unable to retrieve drive Client %v", err)
	}

	
	folderId := "1j2_le4GUrCGFAvWs_48nUs4GWR0EuLnD"

	files, errr := createFile(srv, f.Name(), "application/octet-stream", f, folderId)

	if errr != nil {
		panic(fmt.Sprintf("Could not create file: %v\n", err))
	}

	fmt.Printf("File '%s' successfully uploaded", files.Name)
	fmt.Printf("\nFile Id: '%s' ", files.Id)

}
