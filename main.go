package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"bytes"
	"encoding/json"
	"time"
)

// Config represents the handler plugin config.
type Config struct {
	sensu.PluginConfig
	AuthHeader    string
	ApiUrl        string
	ApiKey        string
	AccessToken   string
	Namespace     string
	Entity        string
	TrustedCaFile string
	Expire        int
	Reason        string
	InsecureSkipVerify	bool
}

var (
	re          = regexp.MustCompile(`\s+`)
	description = `
    Silence Sensu entities on-demand. It does not perform any validation.
	It simply consumes events and silences entity referenced in the event.
    `
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-silence-entity-handler",
			Short:    re.ReplaceAllString(description, " "),
			Keyspace: "sensu.io/plugins/sensu-silence-entity-handler/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		&sensu.PluginConfigOption{
			Path:      "api-url",
			Env:       "SENSU_API_URL",
			Argument:  "api-url",
			Shorthand: "",
			Default:   "http://127.0.0.1:8080",
			Usage:     "Sensu API URL",
			Value:     &plugin.ApiUrl,
		},
		&sensu.PluginConfigOption{
			Path:      "api-key",
			Env:       "SENSU_API_KEY",
			Argument:  "api-key",
			Shorthand: "",
			Default:   "",
			Secret:    true,
			Usage:     "Sensu API Key",
			Value:     &plugin.ApiKey,
		},
		&sensu.PluginConfigOption{
			Path:      "access-token",
			Env:       "SENSU_ACCESS_TOKEN",
			Argument:  "access-token",
			Shorthand: "",
			Default:   "",
			Secret:    true,
			Usage:     "Sensu Access Token",
			Value:     &plugin.AccessToken,
		},
		&sensu.PluginConfigOption{
			Path:      "namespace",
			Env:       "SENSU_NAMESPACE",
			Argument:  "namespace",
			Shorthand: "",
			Default:   "",
			Usage:     "Sensu Namespace",
			Value:     &plugin.Namespace,
		},
		&sensu.PluginConfigOption{
			Path:      "trusted-ca-file",
			Env:       "SENSU_TRUSTED_CA_FILE",
			Argument:  "trusted-ca-file",
			Shorthand: "",
			Default:   "",
			Usage:     "Sensu Trusted Certificate Authority file",
			Value:     &plugin.TrustedCaFile,
		},
		&sensu.PluginConfigOption{
			Path:      "insecure-skip-tls-verify",
			Env:       "",
			Argument:  "insecure-skip-tls-verify",
			Shorthand: "i",
			Default:   false,
			Usage:     "skip TLS certificate verification",
			Value:     &plugin.InsecureSkipVerify,
		},
		&sensu.PluginConfigOption{
			Path:      "expire",
			Env:       "",
			Argument:  "expire",
			Shorthand: "e",
			Default:   1800,
			Usage:     "silence period, seconds",
			Value:     &plugin.Expire,
		},
		&sensu.PluginConfigOption{
			Path:      "reason",
			Env:       "",
			Argument:  "reason",
			Shorthand: "r",
			Default:   "sensu-silence-entity-handler",
			Usage:     "Reason",
			Value:     &plugin.Reason,
		},
	}
)

func main() {
	handler := sensu.NewGoHandler(&plugin.PluginConfig, options, checkArgs, executeHandler)
	handler.Execute()
}

func checkArgs(event *types.Event) error {
	plugin.Entity = event.Entity.Name
	if len(plugin.ApiKey) == 0 && len(plugin.AccessToken) == 0 {
		return fmt.Errorf("--api-key or $SENSU_API_KEY, or --access-token or $SENSU_ACCESS_TOKEN environment variable is required!")
	}
	if len(plugin.Namespace) == 0 {
		if len(os.Getenv("SENSU_NAMESPACE")) > 0 {
			plugin.Namespace = os.Getenv("SENSU_NAMESPACE")
		} else {
			plugin.Namespace = event.Entity.Namespace
		}
	}
	if len(plugin.AccessToken) > 0 {
		plugin.AuthHeader = fmt.Sprintf(
			"Bearer %s",
			plugin.AccessToken,
		)
	}
	if len(plugin.ApiKey) > 0 {
		plugin.AuthHeader = fmt.Sprintf(
			"Key %s",
			plugin.ApiKey,
		)
	}
	return nil
}

// LoadCACerts loads the system cert pool.
func LoadCACerts(path string) (*x509.CertPool, error) {
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		log.Printf("ERROR: failed to load system cert pool: %s", err)
		rootCAs = x509.NewCertPool()
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	if path != "" {
		certs, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("ERROR: failed to read CA file (%s): %s", path, err)
			return nil, err
		}
		rootCAs.AppendCertsFromPEM(certs)
	}
	return rootCAs, nil
}

func initHTTPClient() *http.Client {
	certs, err := LoadCACerts(plugin.TrustedCaFile)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
	}
	tlsConfig := &tls.Config{
		RootCAs: certs,
	}
	if plugin.InsecureSkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := &http.Client{
		Transport: tr,
	}
	return client
}

func executeHandler(event *types.Event) error {
	var entityName    	= event.Entity.Name
	var metadata 		= corev2.NewObjectMeta(fmt.Sprintf("entity:%s:%s", entityName, "*"),
											   event.Entity.Namespace)
	var silenced 		= corev2.NewSilenced(metadata)
	silenced.Reason 	= plugin.Reason
	silenced.Expire 	= int64(plugin.Expire)
	silenced.Begin 		= time.Now().Unix()
	silenced.Subscription 	 = fmt.Sprintf("entity:%s", entityName)
	silenced.ExpireOnResolve = false

	if err := silenced.Validate(); err != nil {
		log.Fatalf("silenced validation error: %v\n%s\n", silenced, err)
	}
	payload, err := json.Marshal(silenced)
	if err != nil {
		log.Fatalf("json error: %s\n", err)
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/api/core/v2/namespaces/%s/silenced",
			plugin.ApiUrl,
			plugin.Namespace,
		),
		bytes.NewBuffer(payload),
	)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
	}
	var httpClient *http.Client = initHTTPClient()
	req.Header.Set("Authorization", plugin.AuthHeader)
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
		return err
	} else if resp.StatusCode == 409 {
		log.Fatalf("ERROR: %v %s (%s/%s)\n", resp.StatusCode, http.StatusText(resp.StatusCode), req.URL, metadata.Name)
		return err
		} else if resp.StatusCode == 400 {
			log.Fatalf("ERROR: %v %s\n%s", resp.StatusCode, http.StatusText(resp.StatusCode), payload)
			return err
	} else if resp.StatusCode >= 300 {
		log.Fatalf("ERROR: %v %s", resp.StatusCode, http.StatusText(resp.StatusCode))
		return err
	} else if resp.StatusCode == 201 {
		log.Printf("Successfully silenced entity \"%s\" from namespace \"%s\"", event.Entity.Name, event.Entity.Namespace)
		return nil
	} else {
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("ERROR: %s\n", err)
			return err
		}
		fmt.Printf("%s\n", string(b))
		return nil
	}
	//return nil
}
