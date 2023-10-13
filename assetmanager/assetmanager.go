package assetmanager

import (
	"os"
	"xavagebb/state"
	"xavagebb/types"
)

type AssetTargetType int

func (a AssetTargetType) String() string {
	switch a {
	default:
		panic("invalid asset target type")
	}
}

// info returns the metadata of an asset given path and default path
//
// It is internal, users should be using *Info functions instead
func info(typ, path, defaultPath string) *types.AssetMetadata {
	st, err := os.Stat(state.Config.Meta.CDNPath + "/" + path)

	if err != nil {
		return &types.AssetMetadata{
			DefaultPath: defaultPath,
			Errors:      []string{"File does not exist"},
			Type:        typ,
		}
	}

	if st.IsDir() {
		return &types.AssetMetadata{
			DefaultPath: defaultPath,
			Errors:      []string{"File is a directory"},
			Type:        typ,
		}
	}

	modTime := st.ModTime()

	return &types.AssetMetadata{
		Exists:       true,
		Path:         path,
		DefaultPath:  defaultPath,
		Size:         st.Size(),
		LastModified: &modTime,
		Type:         typ,
	}
}

func AvatarInfo(targetType AssetTargetType, targetId string) *types.AssetMetadata {
	return info("avatar", "avatars/"+targetType.String()+"/"+targetId+".webp", "avatars/default.webp")
}
