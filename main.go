package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return &schema.Provider{
				Schema: map[string]*schema.Schema{
					"user_agent": {
						Type:     schema.TypeString,
						Optional: true,
						Default:  "TerraURL 0.0.1",
					},
				},
				ResourcesMap: map[string]*schema.Resource{
					"terraurl_fetch": &schema.Resource{
						CreateContext: urlFetchCreate,
						ReadContext:   urlFetchRead,
						UpdateContext: urlFetchUpdate,
						DeleteContext: urlFetchDelete,
						Schema: map[string]*schema.Schema{
							"url": {
								Type:     schema.TypeString,
								Required: true,
							},
							"target_path": {
								Type:     schema.TypeString,
								Required: true,
							},
							"last_modified": {
								Type:     schema.TypeString,
								Computed: true,
								ForceNew: true,
							},
							"size": {
								Type:     schema.TypeInt,
								Computed: true,
								ForceNew: true,
							},
						},
					},
				},
				ConfigureContextFunc: configure,
			}
		},
	})
}

// TerraURLClient is our client wrapper with user-agent hax
type TerraURLClient struct {
	userAgent string
	*http.Client
}

func (tc *TerraURLClient) getRequest(ctx context.Context, method, url string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", tc.userAgent)
	return tc.Do(req.WithContext(ctx))
}

func (tc *TerraURLClient) fetchFile(ctx context.Context, d *schema.ResourceData) error {
	resp, err := tc.getRequest(ctx, "GET", d.Get("url").(string))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(d.Get("target_path").(string))
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	return nil
}

func setState(resp *http.Response, d *schema.ResourceData) diag.Diagnostics {
	clen := resp.Header.Get("content-length")
	size, err := strconv.ParseInt(clen, 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}
	d.Set("size", size)
	d.Set("last_modified", resp.Header.Get("last-modified"))

	return nil
}

func configure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return &TerraURLClient{userAgent: d.Get("user_agent").(string), Client: &http.Client{}}, nil
}

func urlFetchCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tc := m.(*TerraURLClient)
	d.SetId(d.Get("url").(string))

	if err := tc.fetchFile(ctx, d); err != nil {
		return diag.FromErr(err)
	}

	return urlFetchRead(ctx, d, m)
}

func urlFetchRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tc := m.(*TerraURLClient)

	resp, err := tc.getRequest(ctx, "HEAD", d.Get("url").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	defer resp.Body.Close()

	if _, err := os.Stat(d.Get("target_path").(string)); err != nil {
		d.SetId("") // delete the resource so it'll be recreated
	}

	return setState(resp, d)
}

func urlFetchUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if d.HasChanges("last_modified", "size", "url", "target_path") {
		if err := m.(*TerraURLClient).fetchFile(ctx, d); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func urlFetchDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := os.Remove(d.Get("target_path").(string)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
