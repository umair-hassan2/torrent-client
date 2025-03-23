package tracker

import (
	"fmt"
	"io"
	"net/http"

	"github.com/umair-hassan2/torrent-client/cmd/torrent_file"
	"github.com/umair-hassan2/torrent-client/pkg/types"
)

type Tracker struct {
	url string
}

func NewTracker(url string) *Tracker {
	return &Tracker{
		url: url,
	}
}

type TrackerResponse struct {
	Interval int
	Peers    []*types.Peer
}

func GetTrackerResponse(trackerUrl string) (*TrackerResponse, error) {
	resp, err := http.Get(trackerUrl)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	parserResponse, err := torrent_file.ParseTrackerResponse(string(body))

	if err != nil {
		return nil, err
	}

	peers, err := parserResponse.GetRemotePeers()
	if err != nil {
		return nil, err
	}

	return NewTrackerResponse(parserResponse.Interval, peers), nil
}

func NewTrackerResponse(interval int, peers []*types.Peer) *TrackerResponse {
	return &TrackerResponse{
		Interval: interval,
		Peers:    peers,
	}
}
