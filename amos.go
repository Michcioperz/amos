package main

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

func ProperFileDuration(file string) (string, error) {
	cmd := exec.Command("ffmpeg", "-i", file, "-f", "null", "-")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}
	cmderr := cmd.Start()
	if cmderr != nil {
		return "", cmderr
	}
	errcont, err := ioutil.ReadAll(stderr)
	if err != nil {
		return "", err
	}
	errtext := string(errcont)
	erri := strings.LastIndex(errtext, "time=")
	if erri == -1 {
		return "", errors.New("time= not found")
	}
	errtext = errtext[erri+5 : erri+16]
	cmderr = cmd.Wait()
	if cmderr != nil {
		return "", cmderr
	}
	return errtext, nil
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	file, header, ferr := r.FormFile("file")
	if ferr != nil {
		log.Print(ferr)
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "it's like, you didn't attach a file? or it was a bad one?")
		return
	}
	defer file.Close()
	ext := path.Ext(header.Filename)
	if ext != ".gif" {
		log.Printf("non-gif extension %v", ext)
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "it's like, you didn't attach a gif? weird stuff")
		return
	}
	tempdir, terr := ioutil.TempDir("", "amostotherescue")
	if terr != nil {
		log.Print(terr, tempdir)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "it's like, I tried to get a dir to work, but it really didn't work, eh")
		return
	}
	fip := path.Join(tempdir, "in.gif")
	fi, fierr := os.Create(fip)
	if fierr != nil {
		log.Print(fierr)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "it's like, I can't save your gif? dunno")
		return
	}
	_, cperr := io.Copy(fi, file)
	if cperr != nil {
		log.Print(fierr)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "it's like, I can't save your gif? dunno")
		return
	}
	fi.Close()
	fop := path.Join(tempdir, "out2.mp4")
	musicdir := "music"
	musics, merr := ioutil.ReadDir(musicdir)
	if merr != nil {
		log.Print(merr)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "it's like, we couldn't find the right music. or any music at all, actually")
		return
	}
	if len(musics) < 1 {
		log.Print("no music")
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "it's like, what if I told you I don't have any music? yeah I checked that drawer too")
		return
	}
	mchoice := musics[rand.Intn(len(musics))].Name()
	dur, durerr := ProperFileDuration(fip)
	if durerr != nil {
		log.Print(durerr)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "it's like, I tried to check how long the gif is, but it didn't add up, yeah")
		return
	}
	log.Print(dur)
	cmd := exec.Command("ffmpeg", "-i", fip, "-i", path.Join(musicdir, mchoice), "-t", dur, fop)
	cmderr := cmd.Start()
	if cmderr != nil {
		log.Print(cmderr)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "it's like, I tried to process, but it didn't process?")
		return
	}
	cmderr = cmd.Wait()
	if cmderr != nil {
		log.Print(cmderr, cmd)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "it's like, it started processing, but it didn't finish? somehow")
		return
	}
	fop2 := path.Join(tempdir, "out.avi")
	cmd = exec.Command("ffmpeg", "-i", fop, fop2)
	cmderr = cmd.Start()
	if cmderr != nil {
		log.Print(cmderr)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "it's like, I tried to process, but it didn't process?")
		return
	}
	cmderr = cmd.Wait()
	if cmderr != nil {
		log.Print(cmderr, cmd)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "it's like, it started processing, but it didn't finish? somehow")
		return
	}
	fop3 := path.Join(tempdir, "out.mp4")
	cmd = exec.Command("ffmpeg", "-i", fop2, "-vf", "scale=trunc(iw/2)*2:trunc(ih/2)*2", fop3)
	cmderr = cmd.Start()
	if cmderr != nil {
		log.Print(cmderr)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "it's like, I tried to process, but it didn't process?")
		return
	}
	cmderr = cmd.Wait()
	if cmderr != nil {
		log.Print(cmderr, cmd)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "it's like, it started processing, but it didn't finish? somehow")
		return
	}
	fo, foerr := os.Open(fop3)
	if foerr != nil {
		log.Print(foerr)
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "it's like, we got it done, but it didn't send back to you, lol")
		return
	}
	w.Header().Add("Content-Disposition", "attachment; filename=\"meme.mp4\"")
	w.Header().Set("Content-Type", "video/mp4")
	w.WriteHeader(http.StatusOK)
	io.Copy(w, fo)
	return
}

func formRender(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func main() {
	http.HandleFunc("/u", uploadHandler)
	http.HandleFunc("/", formRender)
	http.ListenAndServe(":9007", nil)
}
