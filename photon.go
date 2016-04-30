package luminescent

type Photon struct {
	Message  string
	Electron *Electron
}

func CreatePhoton(message string, electron *Electron) *Photon {

	return &Photon{message, electron}

}
