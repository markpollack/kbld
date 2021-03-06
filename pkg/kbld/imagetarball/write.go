package tarball

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"io"
	"sort"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
)

type TarWriter struct {
	tds           *TarDescriptors
	dst           io.Writer
	tf            *tar.Writer
	layersToWrite []ImageLayerTarDescriptor
}

func NewTarWriter(tds *TarDescriptors, dst io.Writer) *TarWriter {
	return &TarWriter{tds, dst, nil, nil}
}

func (w *TarWriter) Write() error {
	w.tf = tar.NewWriter(w.dst)
	defer w.tf.Close()

	tdsBytes, err := json.Marshal(w.tds.tds)
	if err != nil {
		return err
	}

	err = w.writeTarEntry("manifest.json", bytes.NewReader(tdsBytes), int64(len(tdsBytes)))
	if err != nil {
		return err
	}

	for _, td := range w.tds.tds {
		switch {
		case td.Image != nil:
			err := w.writeImage(*td.Image)
			if err != nil {
				return err
			}

		case td.ImageIndex != nil:
			err := w.writeImageIndex(*td.ImageIndex)
			if err != nil {
				return err
			}

		default:
			panic("Unknown item")
		}
	}

	return w.writeLayers()
}

func (w *TarWriter) writeImageIndex(td ImageIndexTarDescriptor) error {
	for _, idx := range td.Indexes {
		err := w.writeImageIndex(idx)
		if err != nil {
			return err
		}
	}

	for _, img := range td.Images {
		err := w.writeImage(img)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *TarWriter) writeImage(td ImageTarDescriptor) error {
	for _, imgLayer := range td.Layers {
		w.layersToWrite = append(w.layersToWrite, imgLayer)
	}
	return nil
}

func (w *TarWriter) writeLayers() error {
	// Sort layers by digest to have deterministic archive
	sort.Slice(w.layersToWrite, func(i, j int) bool {
		return w.layersToWrite[i].Digest < w.layersToWrite[j].Digest
	})

	writtenLayers := map[string]struct{}{}

	for _, imgLayer := range w.layersToWrite {
		digest, err := regv1.NewHash(imgLayer.Digest)
		if err != nil {
			return err
		}

		name := digest.Algorithm + "-" + digest.Hex + ".tar.gz"

		// Dedup layers
		if _, found := writtenLayers[name]; found {
			continue
		}

		stream, err := w.tds.ImageLayerStream(imgLayer)
		if err != nil {
			return err
		}

		err = w.writeTarEntry(name, stream, imgLayer.Size)
		if err != nil {
			return err
		}

		writtenLayers[name] = struct{}{}
	}

	return nil
}

func (w *TarWriter) writeTarEntry(path string, r io.Reader, size int64) error {
	hdr := &tar.Header{
		Mode:     0644,
		Typeflag: tar.TypeReg,
		Size:     size,
		Name:     path,
	}
	if err := w.tf.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := io.Copy(w.tf, r)
	return err
}
