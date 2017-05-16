package main
import (
    "golang.org/x/sys/windows/registry"
    "log"
    "fmt"
    "strings"
)

/*
Pattern keys in windows registry
HKEY_CURRENT_USER\Software\JavaSoft\Prefs\jetbrains\webstorm
HKEY_USERS\S-1-5-21-2089125521-3810783479-3519687698-1001\Software\JavaSoft\Prefs\jetbrains\webstorm
*/

const keyPath = `Software\JavaSoft\Prefs\jetbrains\webstorm`

type Key struct {
    key registry.Key
    path string
    baseKey *Key
}


func NewKey(baseKey *Key, path string) (*Key, error) {
    key, err := registry.OpenKey(baseKey.key, path, registry.ALL_ACCESS)
    if err != nil {
        return nil, err
    }

    return &Key{
        baseKey: baseKey,
        path: path,
        key: key,
    }, nil
}

func (k *Key) Close() {
    if k != nil {
        k.key.Close()
    }
}




func main() {
    //delete for 'HKEY_CURRENT_USER\Software\JavaSoft\Prefs\jetbrains\webstorm'
    currentUserKey := &Key{
        key: registry.CURRENT_USER,
        path: "HKEY_CURRENT_USER",
    }

    webstormUserKey, err := NewKey(currentUserKey, keyPath)
    if err != nil {
        log.Fatal(err)
    }

    deleteEvlsprt(webstormUserKey)
    webstormUserKey.Close()


    //delete for 'HKEY_USERS\S-1-5-21-2089125521-3810783479-3519687698-1001\Software\JavaSoft\Prefs\jetbrains\webstorm'
    //there are many 'HKEY_USERS\bla bla bla\Software\JavaSoft\Prefs\jetbrains\webstorm' as children of 'HKEY_USERS'
    //and we have to search in all of them
    names, err := registry.USERS.ReadSubKeyNames(0)
    if err != nil {
        log.Fatal(err)
    }

    usersKey := &Key{
        key: registry.USERS,
        path: "HKEY_USERS",
    }

    for _, name := range names {
        subKey := fmt.Sprintf("%s\\%s", name, keyPath)
        key, err := NewKey(usersKey, subKey)
        if err != nil {
            log.Printf("Error trying to create a new key for base key '%s' and sub-key '%s' %v", usersKey.path, subKey, err)
            continue
        }
        deleteEvlsprt(key)
        key.Close()
    }


    fmt.Println("Done")
}

func deleteEvlsprt(baseKey *Key) {
    names, _ := baseKey.key.ReadSubKeyNames(0)
    for _, name := range names {
        if strings.HasPrefix(name, "evlsprt") {
            //fmt.Printf("Just deleted the key '%s' for parent key '%s'\n", name, baseKey.path)
            if err := registry.DeleteKey(baseKey.key, name); err != nil {
                log.Printf("Error when deleting the subkey '%s' of the parent key '%s' %v", name, baseKey.path, err)
            }
        } else {
            newBaseKey, err := NewKey(baseKey, name)
            if err != nil {
                log.Printf("Error trying to create a new key for base key '%s' and sub-key '%s' %v", baseKey.path, name, err)
                continue
            }
            deleteEvlsprt(newBaseKey)
            newBaseKey.Close()
        }
    }
}

