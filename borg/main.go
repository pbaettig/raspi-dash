package borg

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"time"
)

var (
	DocumentsRepo = Repo{
		Name:       "documents",
		ID:         "d39u3wpc@d39u3wpc.repo.borgbase.com:repo",
		Passphrase: "borg",
	}

	PhotosRepo = Repo{
		Name:       "photos",
		ID:         "cs0l58ko@cs0l58ko.repo.borgbase.com:repo",
		Passphrase: "borg",
	}
)

type Repo struct {
	Name       string
	ID         string
	Passphrase string
}

type archiveTimestamp time.Time

func (at *archiveTimestamp) UnmarshalJSON(buf []byte) error {
	var s string
	if err := json.Unmarshal(buf, &s); err != nil {
		return err
	}

	ts, err := time.ParseInLocation("2006-01-02T15:04:05.000000", s, time.Local)
	if err != nil {
		return err
	}

	*at = archiveTimestamp(ts)
	return nil
}

func (at archiveTimestamp) String() string {
	return time.Time(at).Format("2.1.2006 15:04:05")
}

type Archive struct {
	// Archive  string `json:"archive"`
	// Barchive string `json:"barchive"`
	ID   string `json:"id"`
	Name string `json:"name"`
	// Start    string `json:"start"`
	Created archiveTimestamp `json:"time"`
	Age     time.Duration
}
type ArchivesTimestampSorted []Archive

// func (a ArchivesTimestampSorted) Len() int      { return len(a) }
// func (a ArchivesTimestampSorted) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
// func (a ArchivesTimestampSorted) Less(i, j int) bool {
// 	return time.Time(a[i].Time).Before(time.Time(a[j].Time))
// }

type archivesJSON struct {
	Archives []Archive `json:"archives"`
}

func (b Repo) ListBackupArchives() ([]Archive, error) {
	cmd := exec.Command("sudo", "--preserve-env=HOME,BORG_REPO,BORG_PASSPHRASE", "-u", "borg", "borg", "list", "--json")
	cmd.Env = append(os.Environ(),
		"HOME=/home/borg",
		fmt.Sprintf("BORG_REPO=%s", b.ID),
		fmt.Sprintf("BORG_PASSPHRASE=%s", b.Passphrase),
	)
	out, err := cmd.CombinedOutput()
	outS := string(out)

	if err != nil {
		return []Archive{}, fmt.Errorf("borg failed: %s", outS)
	}

	archives := new(archivesJSON)
	err = json.Unmarshal(out, archives)
	if err != nil {
		return []Archive{}, err
	}

	sort.Slice(archives.Archives, func(i, j int) bool {
		return time.Time(archives.Archives[i].Created).After(time.Time(archives.Archives[j].Created))
	})

	// sort.Sort(sort.Reverse(ArchivesTimestampSorted(archives.Archives)))
	return archives.Archives, nil
}

func (b Repo) NewestBackupArchive() (Archive, error) {
	as, err := b.ListBackupArchives()
	if err != nil {
		return Archive{}, fmt.Errorf("NewestBackupArchive cannot list backups: %w", err)
	}

	return as[0], nil
}

// #!/bin/bash

// if [[ "$USER" == "borg" ]]; then
// 	borg_cmd='borg'
// else
// 	# Run borg as borg user with sudo if the script was started by a different user
// 	# not sure if this entirely necessary with a remote repo but won't hurt
// 	borg_cmd='sudo --preserve-env=HOME,BORG_REPO,BORG_PASSPHRASE -u borg borg'
// fi

// repo_name="$1"
// shift

// repo_id=''
// if [[ "$repo_name" == "documents" ]]; then
//     repo_id='d39u3wpc@d39u3wpc.repo.borgbase.com:repo'
// fi

// if [[ "$repo_name" == "photos" ]]; then
//     repo_id='cs0l58ko@cs0l58ko.repo.borgbase.com:repo'
// fi

// if [[ "$repo_id" == "" ]]; then
//     echo "USAGE: $0 [documents|photos] ...borg-options"
//     exit 1
// fi

// # ensure the correct env variables are set
// HOME=/home/borg BORG_REPO="$repo_id" BORG_PASSPHRASE=borg $borg_cmd $@
