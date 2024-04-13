package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	authURL    string = `https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/%s:pull`
	manifesURL string = `https://index.docker.io/v2/library/%s/manifests/latest`
	blobURL    string = `https://index.docker.io/v2/library/%s/blobs/%s`
)

func Authenticate(image string) (string, error) {
	resp, err := http.Get(fmt.Sprintf(authURL, image))
	if err != nil && resp.StatusCode != 200 {
		log.Fatalf("Error authenticating on auth.docker.io: %v", err)
		return "", nil
	}
	defer resp.Body.Close()

	var authResponse AuthResponse
	json.NewDecoder(resp.Body).Decode(&authResponse)
	return authResponse.Token, nil
}

func FetchManifest(image string, token string) (*ManifestResponse, error) {
	client := http.Client{}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(manifesURL, image), nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
		return &ManifestResponse{}, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	// for v2 of the response
	req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := client.Do(req)
	if err != nil && resp.StatusCode != 200 {
		log.Fatalf("Error fetching manifest: %v", err)
		return &ManifestResponse{}, err
	}
	defer resp.Body.Close()

	var manifestResponse ManifestResponse
	json.NewDecoder(resp.Body).Decode(&manifestResponse)
	fmt.Println(resp)

	return &manifestResponse, nil
}

func FetchLayers(image string, token string, manifest ManifestResponse) error {
	client := http.Client{}

	for _, layer := range manifest.Layers {
		req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(blobURL, image, layer.Digest), nil)
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
			return err
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		if err != nil && resp.StatusCode != 200 {
			log.Fatalf("Error fetching manifest: %v", err)
			return err
		}
		defer resp.Body.Close()

		// Create the file
		out, err := os.Create("/tmp/dockerium/layer.tar")
		if err != nil {
			return err
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

func FetchConfig(image string, token string, manifest ManifestResponse) (Config, error) {
	client := http.Client{}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(blobURL, image, manifest.Config.Digest), nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
		return Config{}, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := client.Do(req)
	if err != nil && resp.StatusCode != 200 {
		log.Fatalf("Error fetching manifest: %v", err)
		return Config{}, err
	}

	defer resp.Body.Close()
	var config Config
	json.NewDecoder(resp.Body).Decode(&config)
	fmt.Println(&config)

	return config, nil
}

func ExtractLayer(filepath string) error {
	// TODO: complete this function
	return nil
}
