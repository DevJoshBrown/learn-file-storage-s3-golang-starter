package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	//
	const maxMemory int64 = 10 * 1024 * 1024
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to parse multipartForm", err)
		return
	}

	metadata, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to get video", err)
		return
	}

	if metadata.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "User ID does not match video.userID", err)
		return
	}

	file, fileheader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Could not fetch file & fileheader with FromFile from the request", err)
		return
	}
	defer file.Close()

	mediaType, _, err := mime.ParseMediaType(fileheader.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid Content-Type header", err)
		return
	}

	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusUnsupportedMediaType, "invalid mediaType", nil)
		return
	}

	mediaTypeSplit := strings.Split(mediaType, "/")
	fileExt := mediaTypeSplit[1]

	filename := fmt.Sprintf("%s.%s", videoID, fileExt)
	path := filepath.Join(cfg.assetsRoot, filename)

	diskFile, err := os.Create(path)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create file from filepath", err)
		return
	}
	defer diskFile.Close()

	_, err = io.Copy(diskFile, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not copy upload to disk", err)
		return
	}

	thumbnailURL := fmt.Sprintf("http://localhost:%s/assets/%s.%s", cfg.port, videoID, fileExt)
	metadata.ThumbnailURL = &thumbnailURL

	err = cfg.db.UpdateVideo(metadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, metadata)
}
