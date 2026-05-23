package restconf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/freeconf/yang/fc"
)

func SplitAddress(fullurl string) (address string, module string, path string, err error) {
	eoSlashSlash := strings.Index(fullurl, "//") + 2
	if eoSlashSlash < 2 {
		err = ErrBadAddress
		return
	}
	eoSlash := eoSlashSlash + strings.IndexRune(fullurl[eoSlashSlash:], '/') + 1
	if eoSlash <= eoSlashSlash {
		err = ErrBadAddress
		return
	}
	colon := eoSlash + strings.IndexRune(fullurl[eoSlash:], ':')
	if colon <= eoSlash {
		err = ErrBadAddress
		return
	}
	moduleBegin := strings.LastIndex(fullurl[:colon], "/")
	address = fullurl[:moduleBegin+1]
	module = fullurl[moduleBegin+1 : colon]
	path = fullurl[colon+1:]
	return
}

func SplitUri(uri string) (module string, path string, err error) {
	colon := strings.IndexRune(uri, ':')
	if colon < 0 {
		err = ErrBadAddress
		return
	}
	module = uri[:colon]
	if slash := strings.LastIndex(module, "/"); slash >= 0 {
		module = module[slash+1:]
	}
	path = uri[colon+1:]
	return
}

func FindDeviceIdInUrl(addr string) string {
	segs := strings.SplitAfter(addr, "/restconf=")
	if len(segs) == 2 {
		post := segs[1]
		return post[:len(post)-1]
	}
	return ""
}

func handleErr(compliance ComplianceOptions, err error, r *http.Request, w http.ResponseWriter, mime MimeType) bool {
	if err == nil {
		return false
	}
	fc.Debug.Printf("web request error [%s] %s %s", r.Method, r.URL, err.Error())
	msg := err.Error()
	code := fc.HttpStatusCode(err)
	if !compliance.SimpleErrorResponse {
		errResp := errResponse{
			Type:    "protocol",
			Tag:     decodeErrorTag(code, err),
			Path:    decodeErrorPath(r.RequestURI),
			Message: msg,
		}
		var buff bytes.Buffer
		emsg := map[string]interface{}{
			"ietf-restconf:errors": map[string]interface{}{
				"error": []errResponse{errResp},
			},
		}
		if eerr := json.NewEncoder(&buff).Encode(emsg); eerr != nil {
			fc.Err.Printf("error encoding json error response %s", eerr)
		}
		msg = buff.String()
	}
	w.Header().Set("Content-Type", string(mime))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	fmt.Fprintln(w, msg)
	return true
}

func decodeErrorTag(code int, _err error) string {
	switch code {
	case 409:
		return "in-use"
	case 400:
		return "invalid-value"
	case 401:
		return "access-denied"
	}
	return "operation-failed"
}

func decodeErrorPath(fullPath string) string {
	module, path, err := SplitUri(fullPath)
	if err != nil {
		return fullPath
	}
	return fmt.Sprint(module, ":", path)
}

type errResponse struct {
	Type    string `json:"error-type"`
	Tag     string `json:"error-tag"`
	Path    string `json:"error-path"`
	Message string `json:"error-message"`
}

func ipAddrSplitHostPort(addr string) (host string, port string) {
	bracket := strings.IndexRune(addr, ']')
	dblColon := strings.Index(addr, "::")
	isIpv6 := (bracket >= 0 || dblColon >= 0)
	if isIpv6 {
		if bracket > 0 {
			host = addr[:bracket+1]
			if len(addr) > bracket+2 {
				port = addr[bracket+2:]
			}
		} else {
			host = addr
		}
	} else {
		colon := strings.IndexRune(addr, ':')
		if colon < 0 {
			host = addr
		} else {
			host = addr[:colon]
			port = addr[colon+1:]
		}
	}
	return
}

func appendUrlSegment(a string, b string) string {
	if a == "" || b == "" {
		return a + b
	}
	slashA := a[len(a)-1] == '/'
	slashB := b[0] == '/'
	if slashA != slashB {
		return a + b
	}
	if slashA && slashB {
		return a + b[1:]
	}
	return a + "/" + b
}

func shift(orig *url.URL, delim rune) (string, *url.URL) {
	if orig.Path == "" {
		return "", orig
	}
	copy := *orig
	var segment string
	segment, copy.Path = shiftInString(copy.Path, delim)
	_, copy.RawPath = shiftInString(copy.RawPath, delim)
	return segment, &copy
}

func shiftInString(orig string, delim rune) (string, string) {
	termPos := strings.IndexRune(orig, delim)
	if termPos == 0 {
		orig = orig[1:]
		termPos = strings.IndexRune(orig, delim)
	}
	var shifted string
	var segment string
	if termPos < 0 {
		segment = orig
	} else {
		segment = orig[:termPos]
		shifted = orig[termPos+1:]
	}
	return segment, shifted
}

func shiftOptionalParamWithinSegment(orig *url.URL, optionalDelim rune, segDelim rune) (string, string, *url.URL) {
	copy := *orig
	var segment, optional string
	segment, optional, copy.Path = shiftOptionalParamWithinSegmentInString(copy.Path, optionalDelim, segDelim)
	_, _, copy.RawPath = shiftOptionalParamWithinSegmentInString(copy.RawPath, optionalDelim, segDelim)
	return segment, optional, &copy
}

func shiftOptionalParamWithinSegmentInString(orig string, optionalDelim rune, segDelim rune) (string, string, string) {
	termPos := strings.IndexRune(orig, segDelim)
	if termPos == 0 {
		orig = orig[1:]
		termPos = strings.IndexRune(orig, segDelim)
	}
	var shifted string
	var segment string
	if termPos < 0 {
		segment = orig
	} else {
		segment = orig[:termPos]
		shifted = orig[termPos+1:]
	}
	optPos := strings.IndexRune(segment, optionalDelim)
	if optPos < 0 {
		return segment, "", shifted
	}
	var optional string
	if len(segment) > optPos+1 {
		optional = segment[optPos+1:]
	}
	segment = segment[:optPos]
	return segment, optional, shifted
}