// Copyright IBM Corp. 2021 All Rights Reserved.
// Licensed under the Mozilla Public License v2.0

package ibm

import (
	"context"
	"fmt"
	"log"

	"github.com/IBM-Cloud/container-services-go-sdk/kubernetesserviceapiv1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceIbmContainerNlbDns() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIbmContainerNlbDnsCreate,
		ReadContext:   resourceIbmContainerNlbDnsRead,
		UpdateContext: resourceIbmContainerNlbDnsUpdate,
		DeleteContext: resourceIbmContainerNlbDnsDelete,

		Schema: map[string]*schema.Schema{
			"cluster": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name or ID of the cluster. To list the clusters that you have access to, use the `GET /v1/clusters` API or run `ibmcloud ks cluster ls`.",
			},
			"nlb_host": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"nlb_ips": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"nlb_dns_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"nlb_monitor_state": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"nlb_ssl_secret_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"nlb_ssl_secret_status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"nlb_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret_namespace": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_group_id": &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: applyOnce,
				Description:      "The ID of the resource group that the cluster is in. To check the resource group ID of the cluster, use the GET /v1/clusters/idOrName API. To list available resource group IDs, run ibmcloud resource groups.",
			},
		},
	}
}

func resourceIbmContainerNlbDnsCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	satClient, err := meta.(ClientSession).SatelliteClientSession()
	if err != nil {
		return diag.FromErr(err)
	}

	registerDNSWithIPOptions := &kubernetesserviceapiv1.UpdateDNSWithIPOptions{}
	registerDNSWithIPOptions.SetIdOrName(d.Get("cluster").(string))

	if res, ok := d.GetOk("resource_group_id"); ok {
		header := map[string]string{}
		header["X-Auth-Resource-Group"] = res.(string)
		registerDNSWithIPOptions.SetHeaders(header)
	}
	if _, ok := d.GetOk("nlb_host"); ok {
		registerDNSWithIPOptions.SetNlbHost(d.Get("nlb_host").(string))
	}
	if _, ok := d.GetOk("nlb_ips"); ok {
		ips := []string{}
		for _, segmentsItem := range d.Get("nlb_ips").(*schema.Set).List() {
			ips = append(ips, segmentsItem.(string))
		}
		registerDNSWithIPOptions.SetNlbIPArray(ips)
	}
	response, err := satClient.UpdateDNSWithIPWithContext(context, registerDNSWithIPOptions)
	if err != nil {
		log.Printf("[DEBUG] RegisterDNSWithIPWithContext failed %s\n%s", err, response)
		return diag.FromErr(fmt.Errorf("RegisterDNSWithIPWithContext failed %s\n%s", err, response))
	}

	d.SetId(*registerDNSWithIPOptions.IdOrName)

	return resourceIbmContainerNlbDnsRead(context, d, meta)
}

func resourceIbmContainerNlbDnsRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	kubeClient, err := meta.(ClientSession).VpcContainerAPI()
	if err != nil {
		return diag.FromErr(err)
	}

	nlbData, err := kubeClient.NlbDns().GetNLBDNSList(d.Id())
	if err != nil || nlbData == nil || len(nlbData) < 1 {
		return diag.FromErr(fmt.Errorf("[ERROR] Error Listing NLB DNS (%s): %s", name, err))
	}

	if nlbData != nil {
		for _, nlbConfig := range nlbData {
			if err = d.Set("cluster", d.Id()); err != nil {
				return diag.FromErr(fmt.Errorf("Error setting cluster: %s", err))
			}
			if err = d.Set("nlb_dns_type", nlbConfig.Nlb.DnsType); err != nil {
				return diag.FromErr(fmt.Errorf("Error setting nlb_dns_type: %s", err))
			}
			if err = d.Set("nlb_host", nlbConfig.Nlb.NlbSubdomain); err != nil {
				return diag.FromErr(fmt.Errorf("Error setting nlb_host: %s", err))
			}
			if err = d.Set("nlb_ssl_secret_name", nlbConfig.SecretName); err != nil {
				return diag.FromErr(fmt.Errorf("Error setting nlb_ssl_secret_name: %s", err))
			}
			if err = d.Set("nlb_ssl_secret_status", nlbConfig.SecretStatus); err != nil {
				return diag.FromErr(fmt.Errorf("Error setting nlb_ssl_secret_status: %s", err))
			}
			if err = d.Set("nlb_type", nlbConfig.Nlb.Type); err != nil {
				return diag.FromErr(fmt.Errorf("Error setting nlb_type: %s", err))
			}
			if err = d.Set("secret_namespace", nlbConfig.Nlb.SecretNamespace); err != nil {
				return diag.FromErr(fmt.Errorf("Error setting secret_namespace: %s", err))
			}
		}
	}

	return nil
}

func resourceIbmContainerNlbDnsUpdate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	satClient, err := meta.(ClientSession).SatelliteClientSession()
	if err != nil {
		return diag.FromErr(err)
	}

	updateDNSWithIPOptions := &kubernetesserviceapiv1.UpdateDNSWithIPOptions{}

	updateDNSWithIPOptions.SetIdOrName(d.Id())
	nlbHost := d.Get("nlb_host").(string)

	updateDNSWithIPOptions.NlbHost = ptrToString(nlbHost)

	if d.HasChange("nlb_ips") {
		var remove, add []string
		o, n := d.GetChange("nlb_ips")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		remove = expandStringList(os.Difference(ns).List())
		add = expandStringList(ns.Difference(os).List())

		if len(remove) > 0 {
			unregisterDNSWithIPOptions := &kubernetesserviceapiv1.UnregisterDNSWithIPOptions{}
			unregisterDNSWithIPOptions.SetIdOrName(d.Id())
			unregisterDNSWithIPOptions.SetNlbHost(nlbHost)
			for _, r := range remove {
				unregisterDNSWithIPOptions.SetNlbIP(r)
				response, err := satClient.UnregisterDNSWithIPWithContext(context, unregisterDNSWithIPOptions)
				if err != nil {
					log.Printf("[DEBUG] UnregisterDNSWithIPWithContext failed %s\n%s", err, response)
					return diag.FromErr(fmt.Errorf("UnregisterDNSWithIPWithContext failed %s\n%s", err, response))
				}
			}
		}

		if len(add) > 0 {
			if res, ok := d.GetOk("resource_group_id"); ok {
				header := map[string]string{}
				header["X-Auth-Resource-Group"] = res.(string)
				updateDNSWithIPOptions.SetHeaders(header)
			}
			updateDNSWithIPOptions.SetNlbIPArray(add)
			response, err := satClient.UpdateDNSWithIPWithContext(context, updateDNSWithIPOptions)
			if err != nil {
				log.Printf("[DEBUG] RegisterDNSWithIPWithContext failed %s\n%s", err, response)
				return diag.FromErr(fmt.Errorf("RegisterDNSWithIPWithContext failed %s\n%s", err, response))
			}
		}
	}

	return resourceIbmContainerNlbDnsRead(context, d, meta)
}

func resourceIbmContainerNlbDnsDelete(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	satClient, err := meta.(ClientSession).SatelliteClientSession()
	if err != nil {
		return diag.FromErr(err)
	}

	unregisterDNSWithIPOptions := &kubernetesserviceapiv1.UnregisterDNSWithIPOptions{}
	unregisterDNSWithIPOptions.SetIdOrName(d.Id())
	if res, ok := d.GetOk("resource_group_id"); ok {
		header := map[string]string{}
		header["X-Auth-Resource-Group"] = res.(string)
		unregisterDNSWithIPOptions.SetHeaders(header)
	}
	if nlbHost, ok := d.GetOk("nlb_host"); ok && nlbHost != nil {
		unregisterDNSWithIPOptions.SetNlbHost(nlbHost.(string))
	}

	if ips, ok := d.GetOk("nlb_ips"); ok && ips != nil {
		for _, i := range ips.(*schema.Set).List() {
			unregisterDNSWithIPOptions.SetNlbIP(i.(string))
			response, err := satClient.UnregisterDNSWithIPWithContext(context, unregisterDNSWithIPOptions)
			if err != nil {
				log.Printf("[DEBUG] UnregisterDNSWithIPWithContext failed %s\n%s", err, response)
				return diag.FromErr(fmt.Errorf("UnregisterDNSWithIPWithContext failed %s\n%s", err, response))
			}
		}
	}

	d.SetId("")

	return nil
}
