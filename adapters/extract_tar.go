package adapters

import (
	"archive/tar"
	"context"
	"go.uber.org/zap"
	"io"
	"webploy-server/deployment"
)

func ExtractTarAdapter(ctx context.Context, logger *zap.Logger, d deployment.Deployment, bodyStream io.Reader) ([]string, error) {
	// TODO: Max upload count is handled wrongly:
	// The number of concurrent uploads are tracked by the deployment itself... this is problematic, because
	// it is incremented-decremented on every call of AddFile, when concurrent upload number is high, a new upload may break an already in-progress tar upload...

	var filenames []string // needed for proper auditing
	tr := tar.NewReader(bodyStream)
	for ctx.Err() == nil {
		header, err := tr.Next()

		switch {
		case err == io.EOF:
			logger.Debug("TAR stream ended.", zap.Int("filesCount", len(filenames)))
			return filenames, nil // we are done

		case err != nil:
			logger.Error("Error while reading the next header from TAR stream", zap.Error(err))
			return filenames, err // something went wrong

		case header == nil:
			logger.Debug("Invalid header? Skipping...")
			continue // lol?

		}

		if header.Typeflag != tar.TypeReg {
			// we only allow regular files to be created. Not even directories (we have our own directory creating mechanism)
			logger.Debug("The TAR stream contains un-allowed entry. Ignoring...", zap.Uint8("typeFlag", header.Typeflag), zap.String("name", header.Name))
			continue
		}

		err = d.AddFile(ctx, header.Name, io.NopCloser(tr))
		if err != nil {
			logger.Error("Failed to add file to the deployment from tar stream", zap.Error(err))
			return filenames, err
		}

		filenames = append(filenames, header.Name)
	}

	return filenames, ctx.Err()
}
