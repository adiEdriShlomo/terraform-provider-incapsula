package incapsula

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceIncapRule() *schema.Resource {
	return &schema.Resource{
		Create: resourceIncapRuleCreate,
		Read:   resourceIncapRuleRead,
		Update: resourceIncapRuleUpdate,
		Delete: resourceIncapRuleDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idSlice := strings.Split(d.Id(), "/")
				if len(idSlice) != 2 || idSlice[0] == "" || idSlice[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected site_id/rule_id", d.Id())
				}

				siteID, err := strconv.Atoi(idSlice[0])
				ruleID := idSlice[1]
				if err != nil {
					return nil, err
				}

				d.Set("site_id", siteID)
				d.SetId(ruleID)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			// Required Arguments
			"priority": {
				Description: "The priority",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "The name",
				Type:        schema.TypeString,
				Required:    true,
			},
			"site_id": {
				Description: "Numeric identifier of the site to operate on.",
				Type:        schema.TypeString,
				Required:    true,
			},

			// Optional Arguments
			"enabled": {
				Description: "Is enabled",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "true",
			},
			"action": {
				Description: "The action",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"filter": {
				Description: "The filter",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"allow_caching": {
				Description: "Allow caching",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"dc_id": {
				Description: "The data center id",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"from": {
				Description: "From",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"to": {
				Description: "To",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"response_code": {
				Description: "The response code",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"add_missing": {
				Description: "Add missing",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"rewrite_name": {
				Description: "Rewrite name",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

func getStringValue(key interface{}) string {
	if key == nil {
		return ""
	}
	return key.(string)
}

func resourceIncapRuleCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	addIncapRuleResponse, err := client.AddIncapRule(
		getStringValue(d.Get("enabled")),
		getStringValue(d.Get("name")),
		getStringValue(d.Get("action")),
		getStringValue(d.Get("filter")),
		getStringValue(d.Get("site_id")),
		getStringValue(d.Get("priority")),
		getStringValue(d.Id()),
		getStringValue(d.Get("dc_id")),
		getStringValue(d.Get("allow_caching")),
		getStringValue(d.Get("response_code")),
		getStringValue(d.Get("from")),
		getStringValue(d.Get("to")),
		getStringValue(d.Get("add_missing")),
		getStringValue(d.Get("rewrite_name")),
	)

	if err != nil {
		return err
	}

	d.SetId(addIncapRuleResponse.RuleID)

	return resourceIncapRuleRead(d, m)
}

func resourceIncapRuleRead(d *schema.ResourceData, m interface{}) error {
	// Implement by reading the SiteResponse for the site
	client := m.(*Client)

	var includeAdRules = ""
	var includeIncapRules = ""
	switch action := getStringValue(d.Get("action")); action {
	case "":
		return nil
	case actionAlert:
		fallthrough
	case actionBlockIP:
		fallthrough
	case actionBlockRequest:
		fallthrough
	case actionBlockSession:
		fallthrough
	case actionCaptcha:
		fallthrough
	case actionRetry:
		fallthrough
	case actionIntrusiveHtml:
		includeAdRules = "No"
		includeIncapRules = "Yes"
	case actionDeleteCookie:
		fallthrough
	case actionDeleteHeader:
		fallthrough
	case actionFwdToDataCenter:
		fallthrough
	case actionRedirect:
		fallthrough
	case actionRewriteCookie:
		fallthrough
	case actionRewriteHeader:
		fallthrough
	case actionRewriteUrl:
		includeAdRules = "Yes"
		includeIncapRules = "No"
	}

	listIncapRulesResponse, err := client.ListIncapRules(
		d.Get("site_id").(string),
		includeAdRules,
		includeIncapRules,
	)

	for _, incapRule := range listIncapRulesResponse.IncapRules.All {
		ruleID := d.Id()
		if incapRule.ID == ruleID {
			d.Set("include_ad_rules", includeAdRules)
			d.Set("include_incap_rules", includeIncapRules)
			break
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func resourceIncapRuleUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceIncapRuleCreate(d, m)
}

func resourceIncapRuleDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*Client)

	accountResponse, accountErr := client.Verify()

	if accountErr != nil {
		return accountErr
	}

	accountID := strconv.Itoa(accountResponse.AccountID)
	// Implement delete by clearing out the rule configuration
	err := client.DeleteIncapRule(d.Id(), accountID)

	if err != nil {
		return err
	}

	// Set the ID to empty
	// Implicitly clears the resource
	d.SetId("")

	return nil
}
