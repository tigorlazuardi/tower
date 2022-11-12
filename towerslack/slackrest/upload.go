package slackrest

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"syscall"
)

type FilesUploadPayload struct {
	Channels       []string
	Content        string
	File           io.Reader
	Filename       string
	Filetype       string
	InitialComment string
	ThreadTS       string
	Title          string
}

func (f FilesUploadPayload) FormData(writer io.Writer) error {
	w := multipart.NewWriter(writer)
	defer w.Close()
	if f.Content != "" {
		err := w.WriteField("content", f.Content)
		if err != nil {
			return err
		}
	}
	if f.Filename != "" {
		err := w.WriteField("filename", f.Filename)
		if err != nil {
			return err
		}
	}
	if f.Filetype != "" {
		err := w.WriteField("filetype", f.Filetype)
		if err != nil {
			return err
		}
	}
	if f.InitialComment != "" {
		err := w.WriteField("initial_comment", f.InitialComment)
		if err != nil {
			return err
		}
	}
	if f.ThreadTS != "" {
		err := w.WriteField("thread_ts", f.ThreadTS)
		if err != nil {
			return err
		}
	}
	if f.Title != "" {
		err := w.WriteField("title", f.Title)
		if err != nil {
			return err
		}
	}

	if f.File != nil {
		formFile, err := w.CreateFormField("file")
		if err != nil {
			return err
		}
		_, err = io.Copy(formFile, f.File)
		if errors.Is(err, io.EOF) {
			return nil
		} else if errors.Is(err, syscall.EPIPE) {
			return nil
		} else if err != nil {
			return err
		}
	}
	return nil
}

type FileThread struct {
	ReplyUsers      []interface{} `json:"reply_users"`
	ReplyUsersCount int           `json:"reply_users_count"`
	ReplyCount      int           `json:"reply_count"`
	Ts              string        `json:"ts"`
}

type FileShares struct {
	Private map[string][]FileThread `json:"private"`
	Public  map[string][]FileThread `json:"public"`
}

type FileResponse struct {
	Id                 string     `json:"id"`
	Created            int        `json:"created"`
	Timestamp          int        `json:"timestamp"`
	Name               string     `json:"name"`
	Title              string     `json:"title"`
	Mimetype           string     `json:"mimetype"`
	Filetype           string     `json:"filetype"`
	PrettyType         string     `json:"pretty_type"`
	User               string     `json:"user"`
	Editable           bool       `json:"editable"`
	Size               int        `json:"size"`
	Mode               string     `json:"mode"`
	IsExternal         bool       `json:"is_external"`
	ExternalType       string     `json:"external_type"`
	IsPublic           bool       `json:"is_public"`
	PublicUrlShared    bool       `json:"public_url_shared"`
	DisplayAsBot       bool       `json:"display_as_bot"`
	Username           string     `json:"username"`
	UrlPrivate         string     `json:"url_private"`
	UrlPrivateDownload string     `json:"url_private_download"`
	Thumb64            string     `json:"thumb_64"`
	Thumb80            string     `json:"thumb_80"`
	Thumb360           string     `json:"thumb_360"`
	Thumb360W          int        `json:"thumb_360_w"`
	Thumb360H          int        `json:"thumb_360_h"`
	Thumb480           string     `json:"thumb_480"`
	Thumb480W          int        `json:"thumb_480_w"`
	Thumb480H          int        `json:"thumb_480_h"`
	Thumb160           string     `json:"thumb_160"`
	ImageExifRotation  int        `json:"image_exif_rotation"`
	OriginalW          int        `json:"original_w"`
	OriginalH          int        `json:"original_h"`
	Permalink          string     `json:"permalink"`
	PermalinkPublic    string     `json:"permalink_public"`
	CommentsCount      int        `json:"comments_count"`
	IsStarred          bool       `json:"is_starred"`
	Shares             FileShares `json:"shares"`
	Channels           []string   `json:"channels"`
	Groups             []string   `json:"groups"`
	Ims                []string   `json:"ims"`
	HasRichPreview     bool       `json:"has_rich_preview"`
}

type FilesUploadResponse struct {
	Ok   bool         `json:"ok"`
	File FileResponse `json:"file"`
}

// FileUpload uploads file to slack.
func FileUpload(ctx context.Context, client Client, payload FilesUploadPayload) (resp FilesUploadResponse, err error) {

	return
}
