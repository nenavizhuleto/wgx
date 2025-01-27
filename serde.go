package wgx

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
)

type SerdeJSON struct {
	name string
	path string
}

func NewSerdeJSON(path, name string) *SerdeJSON {
	return &SerdeJSON{
		name: name,
		path: path,
	}
}

func (sd SerdeJSON) Serialize(s Server) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(
		path.Join(sd.path, JoinNameExt(sd.name, JsonExt)),
		data,
		0666,
	); err != nil {
		return err
	}

	return nil
}

func (sd SerdeJSON) Deserialize(s *Server) error {
	data, err := os.ReadFile(path.Join(sd.path, JoinNameExt(sd.name, JsonExt)))
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, s); err != nil {
		return err
	}

	return nil
}

const ConfigurationTemplate = `
# Note: Do not edit this file directly.
# Your changes will be overwritten!

# Server
[Interface]
PrivateKey	= %s
Address		= %s/24
ListenPort	= %s
PreUp		= %s
PostUp		= %s
PreDown		= %s
PostDown	= %s

# Clients
%s

`

const ConfigurationClientTemplate = `
# %s (%s)
[Peer]
PublicKey	= %s
AllowedIPs	= %s/32
`

type SerdeConf struct {
	name string
	path string
}

func NewSerdeConf(path, name string) *SerdeConf {
	return &SerdeConf{
		name: name,
		path: path,
	}
}

func (sd SerdeConf) Serialize(s Server) error {
	var section strings.Builder

	for id, client := range s.Clients {
		entry := fmt.Sprintf(ConfigurationClientTemplate,
			client.Name,
			id.String(),
			client.PublicKey,
			client.Address,
		)

		section.WriteString(entry)
	}

	config := fmt.Sprintf(
		ConfigurationTemplate,
		s.PrivateKey,
		s.Address,
		s.Port,
		"",               // PreUp
		"",               // PostUp
		"",               // PreDown
		"",               // PostDown
		section.String(), // Clients
	)

	if err := os.WriteFile(
		path.Join(sd.path, JoinNameExt(sd.name, ConfExt)),
		[]byte(config),
		0666,
	); err != nil {
		return err
	}

	return nil
}

func (sd SerdeConf) Deserialize(s Server) error {
	panic("unimplemented")
}

func JoinNameExt(name, ext string) string {
	return fmt.Sprintf("%s.%s", name, ext)
}
