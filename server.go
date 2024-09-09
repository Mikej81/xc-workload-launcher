package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/submit", SubmitHandler).Methods("POST")

	http.Handle("/", r)
	fmt.Println("Server starting at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "form.html")
}

func SubmitHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	tenantURL := r.FormValue("tenant_url")
	namespace := r.FormValue("namespace")
	apiToken := r.FormValue("api_token")

	// JSON payload you want to post
	postData := map[string]interface{}{
		"metadata": map[string]interface{}{
			"name":        "demo-nginx-unpriv",
			"namespace":   namespace, // Use the namespace from the form
			"labels":      map[string]interface{}{},
			"annotations": map[string]interface{}{},
			"disable":     false,
		},
		"spec": map[string]interface{}{
			"service": map[string]interface{}{
				"num_replicas": 1,
				"containers": []map[string]interface{}{
					{
						"name": "nginx-container",
						"image": map[string]interface{}{
							"name":        "ghcr.io/nginxinc/nginx-unprivileged",
							"public":      map[string]interface{}{},
							"pull_policy": "IMAGE_PULL_POLICY_DEFAULT",
						},
						"init_container": false,
						"flavor":         "CONTAINER_FLAVOR_TYPE_TINY",
						"command":        []interface{}{},
						"args":           []interface{}{},
					},
				},
				"volumes": []interface{}{},
				"deploy_options": map[string]interface{}{
					"all_res": map[string]interface{}{},
				},
				"advertise_options": map[string]interface{}{
					"advertise_in_cluster": map[string]interface{}{
						"port": map[string]interface{}{
							"info": map[string]interface{}{
								"port":         8080,
								"protocol":     "PROTOCOL_TCP",
								"same_as_port": map[string]interface{}{},
							},
						},
					},
				},
				"family": map[string]interface{}{
					"v4": map[string]interface{}{},
				},
			},
		},
	}

	jsonData, err := json.Marshal(postData)
	if err != nil {
		http.Error(w, "Error converting data to JSON", http.StatusInternalServerError)
		return
	}

	// Constructing the POST URL with the namespace
	postURL := fmt.Sprintf("%s/api/config/namespaces/%s/workloads", tenantURL, namespace)

	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(jsonData))
	if err != nil {
		http.Error(w, "Error creating POST request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "APIToken "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Error making POST request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		successURL := fmt.Sprintf("%s/web/workspaces/distributed-apps/namespaces/%s/applications/virtual_k8s", tenantURL, namespace)
		renderSuccessPage(w, successURL)
	} else {
		http.Error(w, "Error: API responded with status "+resp.Status, http.StatusInternalServerError)
	}
}

func renderSuccessPage(w http.ResponseWriter, successURL string) {
	tmpl := template.Must(template.New("success").Parse(`
    <!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Success</title>
        <style>
            body {
                font-family: Arial, sans-serif;
                background-color: #f4f4f4;
                margin: 0;
                padding: 0;
                display: flex;
                justify-content: center;
                align-items: center;
                height: 100vh;
            }

            .container {
                background-color: #ffffff;
                padding: 20px;
                border-radius: 8px;
                box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
                width: 100%;
                max-width: 500px;
            }

            h1 {
                text-align: center;
                color: #333333;
            }

            a {
                display: block;
                margin-top: 20px;
                padding: 10px;
                background-color: #4CAF50;
                color: white;
                text-align: center;
                text-decoration: none;
                border-radius: 4px;
                transition: background-color 0.3s ease;
            }

            a:hover {
                background-color: #45a049;
            }

        </style>
    </head>
    <body>
        <div class="container">
            <h1>Success!</h1>
            <p>Your data has been submitted successfully.</p>
            <a href="{{.}}">Go to Application</a>
        </div>
    </body>
    </html>
    `))

	tmpl.Execute(w, successURL)
}
