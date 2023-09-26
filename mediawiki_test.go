package mediawiki MEHRDAD Razmi 
test.client.edittokenhttps://github.com/sadbox/mediawikiclient.Domainr.Methodr.MultipartForm.Valuepackage mediawiki

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const (
	editTokenReponse = `{"query":{"pages":{"15580374":{"pageid":15580374,"ns":0,"title":"Main Page","contentmodel":"wikitext","pagelanguage":"en","touched":"2013-11-02T14:08:05Z","lastrevid":574690625,"counter":"","length":6391,"starttimestamp":"2013-11-02T15:24:43Z","edittoken":"+\\","revisions":[{"revid":574690625,"parentid":574690493,"minor":"","user":"Tariqabjotu","timestamp":"2013-09-27T03:10:17Z","comment":"removing unnecessary pipe"}]}}}}`
	firstLogin       = `{"login":{"result":"NeedToken","token":"8f48670ddc7fa9d5fa7e7fa2ae147e80","cookieprefix":"wikidb","sessionid":"927e0d298f6f3b5bb21228803fd9c0eb"}}`
	secondLogin      = `{"login":{"result":"Success","token":"8f48670ddc7fa9d5fa7e7fa2ae147e80","cookieprefix":"wikidb","sessionid":"927e0d298f6f3b5bb21228803fd9c0eb"}}`
	failedLogin      = `{"login":{"result":"ERROR THING","token":"8f48670ddc7fa9d5fa7e7fa2ae147e80","cookieprefix":"wikidb","sessionid":"927e0d298f6f3b5bb21228803fd9c0eb"}}`
	readPage         = `{"query-continue":{"revisions":{"rvcontinue":574690493}},"query":{"pages":{"15580374":{"pageid":15580374,"ns":0,"title":"Main Page","revisions":[{"user":"Tariqabjotu","timestamp":"2013-09-27T03:10:17Z","comment":"removing unnecessary pipe","contentformat":"text/x-wiki","contentmodel":"wikitext","*":"FULL PAGE TEXT"}]}}}}`
	fileURL          = `{"query":{"pages":{"107":{"pageid":107,"ns":6,"title":"File:stuff.pdf","imagerepository":"local","imageinfo":[{"url":"%s","descriptionurl":"TEST"}]}}}}`
	fileURLFailed    = `{"query":{"pages":{"544100":{"pageid":544100,"ns":0,"title":"Asdf"}}}}`
	mwerror          = `{"servedby":"mw1123","error":{"code":"unknown_action","info":"Unrecognized value for parameter 'action': blah"}}`
	editsuccess      = `{"Edit":{"result":"Success","pageid":12,"title":"Talk:Main Page","oldrevid":465,"newrevid":471}}`
	editfailure      = `{"Edit":{"result":"Failure!","pageid":12,"title":"Talk:Main Page","oldrevid":465,"newrevid":471}}`
)

type Test struct {
	ts     *httptest.Server
	client *MWApi
}

func (t *Test) TearDown() {
	t.ts.Close()
}

func BuildUp(response string, t *testing.T) *Test {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, response)
	}))
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error in BuildUp: %s", err)
	}
	return &Test{ts: ts, client: client}
}

func TestGetEditToken(t *testing.T) {
	test := BuildUp(editTokenReponse, t)
	defer test.TearDown()

	err := test.client.GetEditToken()
	if err != nil {
		t.Errorf("Could not get edit token: %s", err.Error())
	} else {
		t.Log("Got edit token")
	}
	if test.client.edittoken == `+\` {
		t.Log("Edit token correct")
	} else {
		t.Errorf("Incorrect edit token: %s", test.client.edittoken)
	}
}

func TestLogin(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		if r.Form.Get("lgtoken") == "" {
			fmt.Fprintln(w, firstLogin)
		} else {
			fmt.Fprintln(w, secondLogin)
		}
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err.Error())
	}
	client.Domain = "asdf"
	err = client.Login("asdf", "asdf")
	if err != nil {
		t.Errorf("Client failed to login: %s", err)
	} else {
		t.Log("Client logged in successfully.")
	}
}

func TestLoginFailedSecondary(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		if r.Form.Get("lgtoken") == "" {
			fmt.Fprintln(w, firstLogin)
		} else {
			fmt.Fprintln(w, failedLogin)
		}
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}
	client.Domain = "asdf"
	err = client.Login("asdf", "asdf")
	if err == nil {
		t.Error("Client logged in successfully. (BUT THIS IS BAD!)")
	} else {
		t.Logf("Client failed to login: %s (BUT THIS IS GOOD!)", err)
	}
}

func TestLoginNoPW(t *testing.T) {
	test := BuildUp(failedLogin, t)
	defer test.TearDown()
	err := test.client.Login("", "")
	if err != nil {
		t.Log("Client refused to login with password.")
	} else {
		t.Error("Client did not a return an error with no password set")
	}
}

func TestLoginFailed(t *testing.T) {
	test := BuildUp(failedLogin, t)
	defer test.TearDown()
	err := test.client.Login("asdf", "JKLa")
	if err != nil {
		t.Logf("Failed to log in: %s", err.Error())
	} else {
		t.Error("Client logged in successfully to incorrect response")
	}
}

func TestPostForm(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}

		if r.Header.Get("user-agent") != "mediawiki (Golang) https://github.com/sadbox/mediawiki TESTING" {
			fmt.Fprintln(w, "USERAGENT")
		} else {
			fmt.Fprintln(w, r.Form.Get("KEY"))
		}
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err.Error())
	}
	value, err := client.postForm(url.Values{"KEY": []string{"VALUE"}})
	if err != nil {
		t.Errorf("Error posting data: %s", err.Error())
	}
	returnValue := strings.TrimSpace(string(value))
	if returnValue == "VALUE" {
		t.Log("Successfully posted to web service.")
	} else if returnValue == "USERAGENT" {
		t.Error("Wrong useragent used!")
	} else {
		t.Error("Did not retrieve right value from web service")
	}
}

func TestAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		if r.Form.Get("KEY") != "VALUE" || r.Form.Get("OTHER KEY") != "OTHER VALUE" || r.Form.Get("format") != "json" {
			fmt.Fprintln(w, `{"status":"FAIL"}`)
		} else {
			fmt.Fprintln(w, `{"status":"PASS"}`)
		}
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err.Error())
	}
	body, err := client.API(map[string]string{"KEY": "VALUE", "OTHER KEY": "OTHER VALUE"})
	if err != nil {
		t.Fatalf("API() returned a non-nil error: %s", err.Error())
	}
	if strings.TrimSpace(string(body)) == `{"status":"PASS"}` {
		t.Log("Successfully posted all information via API() call")
	} else {
		t.Error("Did not get PASS back from test server when posting via API()")
	}
}

func TestRead(t *testing.T) {
	test := BuildUp(readPage, t)
	defer test.TearDown()
	page, err := test.client.Read("TESTING PAGE")
	if err != nil {
		t.Fatalf("Unable to read page: %s", err)
	}
	if page.Revisions[0].Body != "FULL PAGE TEXT" {
		t.Error("Page content not correct")
	}
}

func TestDownload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		if r.Method == "POST" {
			fmt.Fprintln(w, fmt.Sprintf(fileURL, r.Form.Get("titles")))
		} else if r.Method == "GET" {
			fmt.Fprintf(w, `THINGS`)
		}
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err.Error())
	}
	file, err := client.Download(ts.URL)
	if err != nil {
		t.Fatalf("Error downloading file: %s", err.Error())
	}
	defer file.Close()
	returned, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatalf("Error reading downloaded file: %s", err.Error())
	}
	if string(returned) != "THINGS" {
		t.Fatalf("Returned file was not correct")
	}
}

func TestDownloadNoFiles(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		if r.Method == "POST" {
			fmt.Fprintln(w, fileURLFailed)
		} else if r.Method == "GET" {
			fmt.Fprintf(w, `THINGS`)
		}
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}
	_, err = client.Download(ts.URL)
	if err != nil {
		t.Log("Successfully returned error when no files were found", err)
	} else {
		t.Fatal("No error return when no files were found", err)
	}
}

func TestUpload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(int64(10000))
		if err != nil {
			panic(err)
		}
		formValues := r.MultipartForm.Value
		referenceValues := map[string]string{
			"action":   "upload",
			"format":   "json",
			"filename": "test.txt",
			"token":    "ASDFASDF",
		}
	NextKey:
		for key, value := range formValues {
			for innerKey, innerValue := range referenceValues {
				if key == innerKey && value[0] == innerValue {
					continue NextKey
				}
			}
			fmt.Fprintln(w, fmt.Sprintf(`{"upload":{"result":"Value did not match: %s"}}`, key))
			return
		}
		uploadedFile, err := r.MultipartForm.File["file"][0].Open()
		if err != nil {
			fmt.Fprintln(w, `{"upload":{"result":"Error opening uploaded file"}}`)
			return
		}
		defer uploadedFile.Close()
		contents, err := ioutil.ReadAll(uploadedFile)
		if err != nil {
			panic(err)
		}
		if string(contents) != "THIS IS A TEST" {
			fmt.Fprintln(w, `{"upload":{"result":"File contents are not valid"}}`)
			return
		}
		fmt.Fprintln(w, `{"upload":{"result":"Success"}}`)
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err.Error())
	}
	client.edittoken = "ASDFASDF"
	err = client.Upload("test.txt", strings.NewReader("THIS IS A TEST"))
	if err != nil {
		t.Fatalf("Error trying to upload file: %s", err)
	}
}

func TestUploadAutoToken(t *testing.T) {
	test := BuildUp(editTokenReponse, t)
	defer test.TearDown()
	test.client.Upload("stuff", strings.NewReader("stuff"))
	if test.client.edittoken == "" {
		t.Fatalf("Edit token did not get set properly")
	}
}

func TestUploadFailResponse(t *testing.T) {
	test := BuildUp(`{"upload":{"result":"THIS SHOULD BE AN ERROR"}}`, t)
	defer test.TearDown()
	test.client.edittoken = "ASDF"
	err := test.client.Upload("stuff", strings.NewReader("stuff"))
	if err == nil {
		t.Fatal("Did not generate error when upload returned failed result", err)
	} else {
		t.Log("Successfully generated error from failed upload", err)
	}
}

func TestMWError(t *testing.T) {
	test := BuildUp(mwerror, t)
	defer test.TearDown()
	_, err := test.client.Read("DOESN'T MATTER")
	if err != nil {
		t.Log("Properly recieved error:", err)
	} else {
		t.Fatalf("Mediawiki error did not get translated to a go error")
	}
}

func TestEditAutoToken(t *testing.T) {
	test := BuildUp(editTokenReponse, t)
	defer test.TearDown()
	test.client.Edit(map[string]string{"thing": "OTHER THING"})
	if test.client.edittoken == "" {
		t.Fatalf("Edit token did not get set properly")
	}
}

func TestEditSuccess(t *testing.T) {
	test := BuildUp(editsuccess, t)
	defer test.TearDown()
	test.client.edittoken = "asdf"
	err := test.client.Edit(map[string]string{"title": "somepage"})
	if err != nil {
		t.Fatal("Did not get non-error response back", err)
	}
}

func TestEditFail(t *testing.T) {
	test := BuildUp(editfailure, t)
	defer test.TearDown()
	err := test.client.Edit(map[string]string{"title": "somepage"})
	if err == nil {
		t.Fatal("Did not get non-error response back", err)
	}
}

func TestBasicAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Basic Zm9vOmJhcg==" {
			t.Fatalf("Did not recieve correct basic auth header: %s", r.Header.Get(r.Header.Get("Authorization")))
		}
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err.Error())
	}

	client.UseBasicAuth = true
	client.BasicAuthUser = "foo"
	client.BasicAuthPass = "bar"

	t.Log(client)

	_, err = client.API()
	if err != nil {
		t.Fatalf("API() returned a non-nil error: %s", err.Error())
	}
}id>https://www.wikidata.org/w/api.php?action=feedrecentchanges&days=7&feedformat=atom&hidebots=1&hidecategorization=1&limit=50&target=User%3AKosboot&urlversion=1</id>
<title>Wikidata - Changes related to "User:Kosboot" [en]</title>
<link rel="self" type="application/atom+xml" href="https://www.wikidata.org/w/api.php?action=feedrecentchanges&days=7&feedformat=atom&hidebots=1&hidecategorization=1&limit=50&target=User%3AKosboot&urlversion=1"/>
<link rel="alternate" type="text/html" href="https://www.wikidata.org/wiki/Special:RecentChangesLinked"/>
<updated>2023-09-26T09:11:38Z</updated>
<subtitle>Related changes</subtitle>
<generator>MediaWiki 1.41.0-wmf.27</generator>
<entry>
<id>https://www.wikidata.org/w/index.php?title=Wikidata:WikiProject_Music/Statistics/Works_by_form&diff=1980979157&oldid=1976491450</id>
<title>Wikidata:WikiProject Music/Statistics/Works by form</title>
<link rel="alternate" type="text/html" href="https://www.wikidata.org/w/index.php?title=Wikidata:WikiProject_Music/Statistics/Works_by_form&diff=1980979157&oldid=1976491450"/>
<updated>2023-09-24T20:52:42Z</updated>
<summary type="html"><p>Weekly update of property usage stats</p> <a href="//www.wikidata.org/w/index.php?title=Wikidata:WikiProject_Music/Statistics/Works_by_form&amp;diff=1980979157&amp;oldid=1976491450">Show changes</a></summary>
<author>
<name>InteGraalityBot</name>
</author>
</entry>
<entry>
<id>https://www.wikidata.org/w/index.php?title=Wikidata:WikiProject_LD4_Wikidata_Affinity_Group/Affinity_Group_Calls&diff=1979594722&oldid=1930106527</id>
<title>Wikidata:WikiProject LD4 Wikidata Affinity Group/Affinity Group Calls</title>
<link rel="alternate" type="text/html" href="https://www.wikidata.org/w/index.php?title=Wikidata:WikiProject_LD4_Wikidata_Affinity_Group/Affinity_Group_Calls&diff=1979594722&oldid=1930106527"/>
<updated>2023-09-22T18:10:31Z</updated>
<summary type="html"><p><span dir="auto"><span class="autocomment">Upcoming Call</span></span></p> <table style="background-color: #fff; color: #202122;" data-mw="interface"> <col class="diff-marker" /> <col class="diff-content" /> <col class="diff-marker" /> <col class="diff-content" /> <tr class="diff-title" lang="en"> <td colspan="2" style="background-color: #fff; color: #202122; text-align: center;">← Older revision</td> <td colspan="2" style="background-color: #fff; color: #202122; text-align: center;">Revision as of 18:10, 22 September 2023</td> </tr><tr><td colspan="4" class="diff-multi" lang="en">(5 intermediate revisions by the same user not shown)</td></tr><tr> <td colspan="2" class="diff-lineno">Line 9:</td> <td colspan="2" class="diff-lineno">Line 9:</td> </tr> <tr> <td class="diff-marker"></td> <td style="background-color: #f8f9fa; color: #202122; font-size: 88%; border-style: solid; border-width: 1px 1px 1px 4px; border-radius: 0.33em; border-color: #eaecf0; vertical-align: top; white-space: pre-wrap;"><br /></td> <td class="diff-marker"></td> <td style="background-color: #f8f9fa; color: #202122; font-size: 88%; border-style: solid; border-width: 1px 1px 1px 4px; border-radius: 0.33em; border-color: #eaecf0; vertical-align: top; white-space: pre-wrap;"><br /></td> </tr> <tr> <td class="diff-marker"></td> <td style="background-color: #f8f9fa; color: #202122; font-size: 88%; border-style: solid; border-width: 1px 1px 1px 4px; border-radius: 0.33em; border-color: #eaecf0; vertical-align: top; white-space: pre-wrap;"><div>== Upcoming Call ==</div></td> <td class="diff-marker"></td> <td style="background-color: #f8f9fa; color: #202122; font-size: 88%; border-style: solid; border-width: 1px 1px 1px 4px; border-radius: 0.33em; border-color: #eaecf0; vertical-align: top; white-space: pre-wrap;"><div>== Upcoming Call ==</div></td> </tr> <tr> <td colspan="2" class="diff-empty diff-side-deleted"></td> <td class="diff-marker" data-marker="+"></td> <td style="color: #202122; font-size: 88%; border-style: solid; border-width: 1px 1px 1px 4px; border-radius: 0.33em; border-color: #a3d3ff; vertical-align: top; white-space: pre-wrap;"><div>2023-10-13: [https://docs.google.com/document/d/1ni0d5qbFZsCQIBtjJzl9X8V7cwLN01gv1MvZ7zpntYU/edit Lars Willighagen on citation.js.org, Wikidata, and plans for more linked data]</div></td> </tr> <tr> <td class="diff-marker" data-marker="−"></td> <td style="color: #202122; font-size: 88%; border-style: solid; border-width: 1px 1px 1px 4px; border-radius: 0.33em; border-color: #ffe49c; vertical-align: top; white-space: pre-wrap;"><br /></td> <td colspan="2" class="diff-empty diff-side-added"></td> </tr> <tr> <td class="diff-marker"><a class="mw-diff-movedpara-left" title="Paragraph was moved. Click to jump to new location." href="#movedpara_5_0_rhs">&#x26AB;</a></td> <td style="color: #202122; font-size: 88%; border-style: solid; border-width: 1px 1px 1px 4px; border-radius: 0.33em; border-color: #ffe49c; vertical-align: top; white-space: pre-wrap;"><div><a name="movedpara_3_0_lhs"></a><del style="font-weight: bold; text-decoration: none;">We will be joining the LD4 Conference on Thursday, July 13th (10am PST/1pm EST) with a</del> [https://docs.google.com/document/d/16BinQcbUSBJWH8CWxfWuKudUFNKZM-Y6eEHRJeylEBQ/edit<del style="font-weight: bold; text-decoration: none;">''</del> Community Intro and Round 2 of <del style="font-weight: bold; text-decoration: none;">''</del>What's Your Wikidata Passion?<del style="font-weight: bold; text-decoration: none;">''</del>] </div></td> <td colspan="2" class="diff-empty diff-side-added"></td> </tr> <tr> <td class="diff-marker" data-marker="−"></td> <td style="color: #202122; font-size: 88%; border-style: solid; border-width: 1px 1px 1px 4px; border-radius: 0.33em; border-color: #ffe49c; vertical-align: top; white-space: pre-wrap;"><br /></td> <td colspan="2" class="diff-empty diff-side-added"></td> </tr> <tr> <td class="diff-marker" data-marker="−"></td> <td style="color: #202122; font-size: 88%; border-style: solid; border-width: 1px 1px 1px 4px; border-radius: 0.33em; border-color: #ffe49c; vertical-align: top; white-space: pre-wrap;"><div>We are currently taking a break from hosting Tuesday Group Calls. Our regular schedule will resume on October 3, 2023.</div></

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const (
	editTokenReponse = `{"query":{"pages":{"15580374":{"pageid":15580374,"ns":0,"title":"Main Page","contentmodel":"wikitext","pagelanguage":"en","touched":"2013-11-02T14:08:05Z","lastrevid":574690625,"counter":"","length":6391,"starttimestamp":"2013-11-02T15:24:43Z","edittoken":"+\\","revisions":[{"revid":574690625,"parentid":574690493,"minor":"","user":"Tariqabjotu","timestamp":"2013-09-27T03:10:17Z","comment":"removing unnecessary pipe"}]}}}}`
	firstLogin       = `{"login":{"result":"NeedToken","token":"8f48670ddc7fa9d5fa7e7fa2ae147e80","cookieprefix":"wikidb","sessionid":"927e0d298f6f3b5bb21228803fd9c0eb"}}`
	secondLogin      = `{"login":{"result":"Success","token":"8f48670ddc7fa9d5fa7e7fa2ae147e80","cookieprefix":"wikidb","sessionid":"927e0d298f6f3b5bb21228803fd9c0eb"}}`
	failedLogin      = `{"login":{"result":"ERROR THING","token":"8f48670ddc7fa9d5fa7e7fa2ae147e80","cookieprefix":"wikidb","sessionid":"927e0d298f6f3b5bb21228803fd9c0eb"}}`
	readPage         = `{"query-continue":{"revisions":{"rvcontinue":574690493}},"query":{"pages":{"15580374":{"pageid":15580374,"ns":0,"title":"Main Page","revisions":[{"user":"Tariqabjotu","timestamp":"2013-09-27T03:10:17Z","comment":"removing unnecessary pipe","contentformat":"text/x-wiki","contentmodel":"wikitext","*":"FULL PAGE TEXT"}]}}}}`
	fileURL          = `{"query":{"pages":{"107":{"pageid":107,"ns":6,"title":"File:stuff.pdf","imagerepository":"local","imageinfo":[{"url":"%s","descriptionurl":"TEST"}]}}}}`
	fileURLFailed    = `{"query":{"pages":{"544100":{"pageid":544100,"ns":0,"title":"Asdf"}}}}`
	mwerror          = `{"servedby":"mw1123","error":{"code":"unknown_action","info":"Unrecognized value for parameter 'action': blah"}}`
	editsuccess      = `{"Edit":{"result":"Success","pageid":12,"title":"Talk:Main Page","oldrevid":465,"newrevid":471}}`
	editfailure      = `{"Edit":{"result":"Failure!","pageid":12,"title":"Talk:Main Page","oldrevid":465,"newrevid":471}}`
)

type Test struct {
	ts     *httptest.Server
	client *MWApi
}

func (t *Test) TearDown() {
	t.ts.Close()
}

func BuildUp(response string, t *testing.T) *Test {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, response)
	}))
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error in BuildUp: %s", err)
	}
	return &Test{ts: ts, client: client}
}

func TestGetEditToken(t *testing.T) {
	test := BuildUp(editTokenReponse, t)
	defer test.TearDown()

	err := test.client.GetEditToken()
	if err != nil {
		t.Errorf("Could not get edit token: %s", err.Error())
	} else {
		t.Log("Got edit token")
	}
	if test.client.edittoken == `+\` {
		t.Log("Edit token correct")
	} else {
		t.Errorf("Incorrect edit token: %s", test.client.edittoken)
	}
}

func TestLogin(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		if r.Form.Get("lgtoken") == "" {
			fmt.Fprintln(w, firstLogin)
		} else {
			fmt.Fprintln(w, secondLogin)
		}
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err.Error())
	}
	client.Domain = "asdf"
	err = client.Login("asdf", "asdf")
	if err != nil {
		t.Errorf("Client failed to login: %s", err)
	} else {
		t.Log("Client logged in successfully.")
	}
}

func TestLoginFailedSecondary(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		if r.Form.Get("lgtoken") == "" {
			fmt.Fprintln(w, firstLogin)
		} else {
			fmt.Fprintln(w, failedLogin)
		}
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}
	client.Domain = "asdf"
	err = client.Login("asdf", "asdf")
	if err == nil {
		t.Error("Client logged in successfully. (BUT THIS IS BAD!)")
	} else {
		t.Logf("Client failed to login: %s (BUT THIS IS GOOD!)", err)
	}
}

func TestLoginNoPW(t *testing.T) {
	test := BuildUp(failedLogin, t)
	defer test.TearDown()
	err := test.client.Login("", "")
	if err != nil {
		t.Log("Client refused to login with password.")
	} else {
		t.Error("Client did not a return an error with no password set")
	}
}

func TestLoginFailed(t *testing.T) {
	test := BuildUp(failedLogin, t)
	defer test.TearDown()
	err := test.client.Login("asdf", "JKLa")
	if err != nil {
		t.Logf("Failed to log in: %s", err.Error())
	} else {
		t.Error("Client logged in successfully to incorrect response")
	}
}

func TestPostForm(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}

		if r.Header.Get("user-agent") != "mediawiki (Golang) https://github.com/sadbox/mediawiki TESTING" {
			fmt.Fprintln(w, "USERAGENT")
		} else {
			fmt.Fprintln(w, r.Form.Get("KEY"))
		}
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err.Error())
	}
	value, err := client.postForm(url.Values{"KEY": []string{"VALUE"}})
	if err != nil {
		t.Errorf("Error posting data: %s", err.Error())
	}
	returnValue := strings.TrimSpace(string(value))
	if returnValue == "VALUE" {
		t.Log("Successfully posted to web service.")
	} else if returnValue == "USERAGENT" {
		t.Error("Wrong useragent used!")
	} else {
		t.Error("Did not retrieve right value from web service")
	}
}

func TestAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		if r.Form.Get("KEY") != "VALUE" || r.Form.Get("OTHER KEY") != "OTHER VALUE" || r.Form.Get("format") != "json" {
			fmt.Fprintln(w, `{"status":"FAIL"}`)
		} else {
			fmt.Fprintln(w, `{"status":"PASS"}`)
		}
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err.Error())
	}
	body, err := client.API(map[string]string{"KEY": "VALUE", "OTHER KEY": "OTHER VALUE"})
	if err != nil {
		t.Fatalf("API() returned a non-nil error: %s", err.Error())
	}
	if strings.TrimSpace(string(body)) == `{"status":"PASS"}` {
		t.Log("Successfully posted all information via API() call")
	} else {
		t.Error("Did not get PASS back from test server when posting via API()")
	}
}

func TestRead(t *testing.T) {
	test := BuildUp(readPage, t)
	defer test.TearDown()
	page, err := test.client.Read("TESTING PAGE")
	if err != nil {
		t.Fatalf("Unable to read page: %s", err)
	}
	if page.Revisions[0].Body != "FULL PAGE TEXT" {
		t.Error("Page content not correct")
	}
}

func TestDownload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		if r.Method == "POST" {
			fmt.Fprintln(w, fmt.Sprintf(fileURL, r.Form.Get("titles")))
		} else if r.Method == "GET" {
			fmt.Fprintf(w, `THINGS`)
		}
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err.Error())
	}
	file, err := client.Download(ts.URL)
	if err != nil {
		t.Fatalf("Error downloading file: %s", err.Error())
	}
	defer file.Close()
	returned, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatalf("Error reading downloaded file: %s", err.Error())
	}
	if string(returned) != "THINGS" {
		t.Fatalf("Returned file was not correct")
	}
}

func TestDownloadNoFiles(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			panic(err)
		}
		if r.Method == "POST" {
			fmt.Fprintln(w, fileURLFailed)
		} else if r.Method == "GET" {
			fmt.Fprintf(w, `THINGS`)
		}
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err)
	}
	_, err = client.Download(ts.URL)
	if err != nil {
		t.Log("Successfully returned error when no files were found", err)
	} else {
		t.Fatal("No error return when no files were found", err)
	}
}

func TestUpload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(int64(10000))
		if err != nil {
			panic(err)
		}
		formValues := r.MultipartForm.Value
		referenceValues := map[string]string{
			"action":   "upload",
			"format":   "json",
			"filename": "test.txt",
			"token":    "ASDFASDF",
		}
	NextKey:
		for key, value := range formValues {
			for innerKey, innerValue := range referenceValues {
				if key == innerKey && value[0] == innerValue {
					continue NextKey
				}
			}
			fmt.Fprintln(w, fmt.Sprintf(`{"upload":{"result":"Value did not match: %s"}}`, key))
			return
		}
		uploadedFile, err := r.MultipartForm.File["file"][0].Open()
		if err != nil {
			fmt.Fprintln(w, `{"upload":{"result":"Error opening uploaded file"}}`)
			return
		}
		defer uploadedFile.Close()
		contents, err := ioutil.ReadAll(uploadedFile)
		if err != nil {
			panic(err)
		}
		if string(contents) != "THIS IS A TEST" {
			fmt.Fprintln(w, `{"upload":{"result":"File contents are not valid"}}`)
			return
		}
		fmt.Fprintln(w, `{"upload":{"result":"Success"}}`)
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err.Error())
	}
	client.edittoken = "ASDFASDF"
	err = client.Upload("test.txt", strings.NewReader("THIS IS A TEST"))
	if err != nil {
		t.Fatalf("Error trying to upload file: %s", err)
	}
}

func TestUploadAutoToken(t *testing.T) {
	test := BuildUp(editTokenReponse, t)
	defer test.TearDown()
	test.client.Upload("stuff", strings.NewReader("stuff"))
	if test.client.edittoken == "" {
		t.Fatalf("Edit token did not get set properly")
	}
}

func TestUploadFailResponse(t *testing.T) {
	test := BuildUp(`{"upload":{"result":"THIS SHOULD BE AN ERROR"}}`, t)
	defer test.TearDown()
	test.client.edittoken = "ASDF"
	err := test.client.Upload("stuff", strings.NewReader("stuff"))
	if err == nil {
		t.Fatal("Did not generate error when upload returned failed result", err)
	} else {
		t.Log("Successfully generated error from failed upload", err)
	}
}

func TestMWError(t *testing.T) {
	test := BuildUp(mwerror, t)
	defer test.TearDown()
	_, err := test.client.Read("DOESN'T MATTER")
	if err != nil {
		t.Log("Properly recieved error:", err)
	} else {
		t.Fatalf("Mediawiki error did not get translated to a go error")
	}
}

func TestEditAutoToken(t *testing.T) {
	test := BuildUp(editTokenReponse, t)
	defer test.TearDown()
	test.client.Edit(map[string]string{"thing": "OTHER THING"})
	if test.client.edittoken == "" {
		t.Fatalf("Edit token did not get set properly")
	}
}

func TestEditSuccess(t *testing.T) {
	test := BuildUp(editsuccess, t)
	defer test.TearDown()
	test.client.edittoken = "asdf"
	err := test.client.Edit(map[string]string{"title": "somepage"})
	if err != nil {
		t.Fatal("Did not get non-error response back", err)
	}
}

func TestEditFail(t *testing.T) {
	test := BuildUp(editfailure, t)
	defer test.TearDown()
	err := test.client.Edit(map[string]string{"title": "somepage"})
	if err == nil {
		t.Fatal("Did not get non-error response back", err)
	}
}

func TestBasicAuth(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Basic Zm9vOmJhcg==" {
			t.Fatalf("Did not recieve correct basic auth header: %s", r.Header.Get(r.Header.Get("Authorization")))
		}
	}))
	defer ts.Close()
	client, err := New(ts.URL, "TESTING")
	if err != nil {
		t.Fatalf("Error creating client: %s", err.Error())
	}

	client.UseBasicAuth = true
	client.BasicAuthUser = "foo"
	client.BasicAuthPass = "bar"

	t.Log(client)

	_, err = client.API()
	if err != nil {
		t.Fatalf("API() returned a non-nil error: %s", err.Error())
	}
}
