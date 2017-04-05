package torrentclient

import (
	"code.uber.internal/go-common.git/x/log"
	"code.uber.internal/infra/kraken/client/store"
	"code.uber.internal/infra/kraken/configuration"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

// Torrent contains layer info and a pointer to cache to retrieve data
type Torrent struct {
	config    *configuration.Config
	store     *store.LocalFileStore
	name      string
	len       int64
	numPieces int
}

// NewTorrent returns a new LayerStore. Caller should then call either LoadFromDisk or CreateEmptyLayerFile.
func NewTorrent(
	config *configuration.Config,
	store *store.LocalFileStore,
	name string,
	len int64,
	numPieces int) *Torrent {
	return &Torrent{
		config:    config,
		store:     store,
		name:      name,
		len:       len,
		numPieces: numPieces,
	}
}

// Open creates a download torrent
func (tor *Torrent) Open() error {
	// create download file from LocalFileStore
	new, err := tor.store.CreateDownloadFile(tor.name, tor.len)
	if err != nil {
		return err
	}

	// if the download is not new, the torrent is either downloading or already downloaded
	if !new {
		log.Infof("Torrent %s has been created already. No actions taken.", tor.name)
		return nil
	}

	// download is new, set metadata for pieces
	_, err = tor.store.SetDownloadFilePieceStatus(tor.name, []byte{store.PieceClean}, 0, tor.numPieces)
	if err != nil {
		log.Errorf("Error setting metadata for new download %s", tor.name)
		return err
	}

	log.Infof("Successfully created new downalod for %s", tor.name)
	return nil
}

// Piece returns pieceStore of the layer given metainfo
func (tor *Torrent) Piece(p metainfo.Piece) storage.PieceImpl {
	if p.Index() >= tor.numPieces {
		log.WithFields(log.Fields{
			"name":  tor.name,
			"piece": p.Index(),
		}).Error("Piece index out of range")
		return nil
	}

	return &Piece{
		config:    tor.config,
		store:     tor.store,
		name:      tor.name,
		index:     p.Index(),
		numPieces: tor.numPieces,
	}
}

// Close closes the LayerStore
func (tor *Torrent) Close() error {
	return nil
}