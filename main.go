package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

const (
	maxRecurseDepth = 10
)

// VaultStore struct keeps Vault client and path
type VaultStore struct {
	VaultClient *vaultapi.Client
}

const dataField = "data"

var errKeyNotExists = fmt.Errorf("vault key not exists")
var errInvalidValue = fmt.Errorf("vault value is malformed")

// NewVaultStore creates new instance of VaultStore
func NewVaultStore() *VaultStore {

	vaultConfig := *vaultapi.DefaultConfig()
	vaultConfig.ReadEnvironment()
	client, err := vaultapi.NewClient(&vaultConfig)
	if err != nil {
		log.Fatalf("failed to create vault client: %v", err)
	}

	return &VaultStore{
		VaultClient: client,
	}

}

func (store *VaultStore) makeError(action string, key string, err error) error {
	return fmt.Errorf("failed to %s key '%s' from %s: %v", action, key, store, err)
}

// List send GET request to the Vault and list keys
func (store *VaultStore) List(key string) ([]string, error) {
	log.Printf("List key %s", key)

	keys, err := store.list(key)
	if err != nil {
		return nil, store.makeError("list", key, err)
	}

	// remove all trailing slashes
	sanitisedKeys := make([]string, len(keys))
	for i, k := range keys {
		sanitisedKeys[i] = strings.TrimSuffix(k, "/")
	}

	return sanitisedKeys, nil
}

// Exists check if the key exist
func (store *VaultStore) Exists(path string) (bool, error) {
	log.Printf("Head key %s", path)

	secret, err := store.VaultClient.Logical().Read(path)
	if err != nil {
		return false, store.makeError("head", path, err)
	}
	return secret != nil, nil
}

// Get receives key/value from Vault
func (store *VaultStore) Get(path string) ([]byte, error) {
	log.Printf("Get key %s", path)

	secret, err := store.VaultClient.Logical().Read(path)
	if err != nil {
		return nil, store.makeError("get", path, err)
	}
	if secret == nil {
		return nil, store.makeError("get", path, errKeyNotExists)
	}
	if secret.Data[dataField] == nil {
		return nil, store.makeError("get", path, errInvalidValue)
	}

	encodedData := secret.Data[dataField].(string)
	decodedData, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode data: %v", err)
	}

	return decodedData, nil
}

// Set new key/value
func (store *VaultStore) Set(path string, data []byte) error {
	encodedData := base64.StdEncoding.EncodeToString(data)
	log.Printf("Set key %s (len: %d bytes)", path, len(encodedData))

	secretData := make(map[string]interface{})
	secretData[dataField] = encodedData
	_, err := store.VaultClient.Logical().Write(path, secretData)
	if err != nil {
		return store.makeError("set", path, err)
	}

	return nil
}

// Delete secret by key
func (store *VaultStore) Delete(path string) error {
	log.Printf("Delete key %s", path)

	_, err := store.VaultClient.Logical().Delete(path)
	if err != nil {
		path = vaultPath.ReplaceAllString(path, `$1/metadata/$2`)
	}
	_, err = store.VaultClient.Logical().Delete(path)
	if err != nil {
		return store.makeError("delete", path, err)
	}

	return nil
}

// DeleteAll deletes all subkeys
func (store *VaultStore) DeleteAll(path string) error {
	log.Printf("Collecting information: %s", path)

	start := time.Now()

	// list all keys to delete
	keysToDel, err := store.listRecurse(path, 0)
	if err != nil {
		return store.makeError("deleteAll", path, err)
	}

	if len(keysToDel) < 1 {
		return fmt.Errorf("can't find keys to delete: %s", path)
	}

	// Print all found keys and ask user delete or not
	fmt.Println("=================")
	fmt.Println("Found keys:")
	for i, k := range keysToDel {
		fmt.Printf("%d %s\n", i, k)
	}
	var answer string
	fmt.Print("Delete all keys (yes/no): ")
	fmt.Scan(&answer)

	if answer == "yes" || answer == "y" {
		log.Printf("DeleteAll: deleting %d keys", len(keysToDel))
		// delete all listed keys
		for _, k := range keysToDel {
			if err := store.Delete(k); err != nil {
				// warning only if one key failed to delete
				log.Printf("failed to delete %s: %v", k, err)
			}
			// fmt.Printf("store.Delete(%s)\n", k)
		}
	}

	fmt.Printf("Spent: %.2fs\n", time.Since(start).Seconds())

	return nil
}

func (store *VaultStore) listRecurse(path string, depth int) ([]string, error) {
	// sanity check (should not exceed 10 levels)
	if depth > maxRecurseDepth {
		return nil, store.makeError("listRecurse", path, fmt.Errorf("maximum recurse depth reached"))
	}

	// list all keys under the current key
	keys, err := store.list(path)
	if err != nil {
		return nil, store.makeError("listRecurse", path, err)
	}

	// create a new string array to hold flattened keys
	flatKeys := make([]string, 0)
	for _, k := range keys {
		fullKey := fmt.Sprintf("%s%s", path, k)

		if !isDirectory(k) {
			// if not a directory, append to the flatKeys directly
			flatKeys = append(flatKeys, fullKey)
		} else {
			// otherwise, call listRecurse on it and append all returned keys to flatKeys
			subKeys, err := store.listRecurse(fullKey, depth+1)
			if err != nil {
				return nil, store.makeError("listRecurse", path, err)
			}
			flatKeys = append(flatKeys, subKeys...)
		}
	}

	// return flat list of keys
	return flatKeys, nil
}

var vaultPath = regexp.MustCompile(`^([A-Za-z0-9]*)/`)

func (store *VaultStore) list(path string) ([]string, error) {

	secret, err := store.VaultClient.Logical().List(path)
	if err != nil {
		return nil, store.makeError("list", path, err)
	}
	if secret == nil {
		return make([]string, 0), nil
	}
	if len(secret.Data) == 0 {
		path = vaultPath.ReplaceAllString(path, `$1/metadata/$2`)
	}
	secret, err = store.VaultClient.Logical().List(path)
	if err != nil {
		return nil, store.makeError("list", path, err)
	}
	if secret == nil {
		return make([]string, 0), nil
	}
	if secret.Data["keys"] == nil {
		return make([]string, 0), nil
	}

	data := secret.Data["keys"]
	return dataAsList(data)
}

func isDirectory(key string) bool {
	return strings.HasSuffix(key, "/")
}

func dataAsList(data interface{}) ([]string, error) {
	if list, ok := data.([]interface{}); ok {
		keys := make([]string, 0)
		for _, k := range list {
			keys = append(keys, k.(string))
		}
		return keys, nil
	}

	return nil, fmt.Errorf("data is not a list")
}

func main() {
	path := flag.String("path", "", "/secret/path")
	flag.Parse()

	vs := NewVaultStore()

	err := vs.DeleteAll(*path)
	if err != nil {
		log.Fatal(err)
	}
}
