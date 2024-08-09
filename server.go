package wgx

import (
	"errors"
	"net"

	"github.com/google/uuid"
)

var (
	ErrClientNameMissing  = errors.New("missing name")
	ErrClientNotFound     = errors.New("not found")
	ErrNoAddressAvailable = errors.New("no address available")
)

type Server struct {
	Interface Interface `json:"interface"`
	Address   string    `json:"address"`
	Port      string    `json:"port"`

	ConfigPath string `json:"-"`

	PrivateKey PrivateKey           `json:"private_key"`
	PublicKey  PublicKey            `json:"public_key"`
	Clients    map[uuid.UUID]Client `json:"clients"`

	serdeJSON *SerdeJSON
	serdeConf *SerdeConf
}

const (
	DefaultAddress    = "10.0.8.1"
	DefaultPort       = "51820"
	DefaultInterface  = "wg0"
	DefaultConfigPath = "/etc/wireguard"

	JsonExt = "json"
	ConfExt = "conf"
)

type ServerSerde interface {
	Serialize(s Server) error
	Deserialize(s *Server) error
}

func NewServer() (*Server, error) {
	pair, err := GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	return &Server{
		Address:    DefaultAddress,
		Port:       DefaultPort,
		Interface:  DefaultInterface,
		ConfigPath: DefaultConfigPath,

		PrivateKey: pair.PrivateKey,
		PublicKey:  pair.PublicKey,

		Clients: make(map[uuid.UUID]Client),

		serdeJSON: NewSerdeJSON(DefaultConfigPath, DefaultInterface),
		serdeConf: NewSerdeConf(DefaultConfigPath, DefaultInterface),
	}, nil
}

func (s *Server) Listen() error {
	return nil
}

func (s *Server) NextAvailableAddress() (string, error) {
	serverAddr := net.ParseIP(s.Address).To4()

	for i := 2; i < 255; i++ {
		probeAddr := net.IP(
			append(serverAddr[:3], byte(i)),
		).To4()

		for _, client := range s.Clients {
			clientAddr := net.ParseIP(client.Address).To4()

			if probeAddr.Equal(clientAddr) {
				goto next
			}
		}

		return probeAddr.String(), nil
	next:
	}

	return "", ErrNoAddressAvailable
}

func (s *Server) Load() error {
	return s.serdeJSON.Deserialize(s)
}

func (s *Server) Save() error {
	if err := s.serdeJSON.Serialize(*s); err != nil {
		return err
	}

	if err := s.serdeConf.Serialize(*s); err != nil {
		return err
	}

	return nil
}

func (s *Server) Sync() error {
	if err := s.Save(); err != nil {
		return err
	}

	defer func() {
		if err := s.Load(); err != nil {
			panic("failed to load after sync")
		}
	}()

	return SyncConf(s.Interface)
}

type Client struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	Address    string     `json:"address"`
	Enabled    bool       `json:"enabled"`
	PrivateKey PrivateKey `json:"private_key"`
	PublicKey  PublicKey  `json:"public_key"`
}

func (s *Server) GetClients() []Client {
	var clients []Client

	for _, client := range s.Clients {
		clients = append(clients, client)
	}

	return clients
}

func (s *Server) AddClient(name string) (Client, error) {
	if name == "" {
		return Client{}, ErrClientNameMissing
	}

	addr, err := s.NextAvailableAddress()
	if err != nil {
		return Client{}, err
	}

	pair, err := GenerateKeyPair()
	if err != nil {
		return Client{}, err
	}

	client := Client{
		ID:      uuid.New(),
		Name:    name,
		Address: addr,
		Enabled: true,

		PrivateKey: pair.PrivateKey,
		PublicKey:  pair.PublicKey,
	}

	s.Clients[client.ID] = client

	return client, nil
}

func (s *Server) RemoveClient(id uuid.UUID) error {
	_, ok := s.Clients[id]
	if !ok {
		return ErrClientNotFound
	}

	delete(s.Clients, id)

	return nil
}

type UpdateClientParams struct {
	Name string `json:"name"`
}

func (s *Server) UpdateClient(id uuid.UUID, params UpdateClientParams) error {
	client, ok := s.Clients[id]
	if !ok {
		return ErrClientNotFound
	}

	client.Name = params.Name

	s.Clients[id] = client

	return nil
}
