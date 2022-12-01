package role

import (
	"math/rand"
	"time"

	"github.com/kairos-io/kairos/pkg/config"

	providerConfig "github.com/kairos-io/provider-kairos/internal/provider/config"
	service "github.com/mudler/edgevpn/api/client/service"
)

// scheduleRoles assigns roles to nodes. Meant to be called only by leaders
// TODO: HA-Auto.
func scheduleRoles(nodes []string, c *service.RoleConfig, cc *config.Config, pconfig *providerConfig.Config) error {
	rand.Seed(time.Now().Unix())

	// Assign roles to nodes
	unassignedNodes, currentRoles, existsMaster := getRoles(c.Client, nodes)

	c.Logger.Infof("I'm the leader. My UUID is: %s.\n Current assigned roles: %+v", c.UUID, currentRoles)
	c.Logger.Infof("Master already present: %t", existsMaster)
	c.Logger.Infof("Unassigned nodes: %+v", unassignedNodes)

	if !existsMaster && len(unassignedNodes) > 0 {
		var selected string
		toSelect := unassignedNodes

		// Avoid to schedule to ourselves if we have a static role
		if pconfig.Kairos.Role != "" {
			toSelect = []string{}
			for _, u := range unassignedNodes {
				if u != c.UUID {
					toSelect = append(toSelect, u)
				}
			}
		}

		// select one node without roles to become master
		if len(toSelect) == 1 {
			selected = toSelect[0]
		} else {
			selected = toSelect[rand.Intn(len(toSelect)-1)]
		}

		if err := c.Client.Set("role", selected, "master"); err != nil {
			return err
		}
		c.Logger.Info("-> Set master to", selected)
		currentRoles[selected] = "master"
		// Return here, so next time we get called
		// makes sure master is set.
		return nil
	}

	// cycle all empty roles and assign worker roles
	for _, uuid := range unassignedNodes {
		if err := c.Client.Set("role", uuid, "worker"); err != nil {
			c.Logger.Error(err)
			return err
		}
		c.Logger.Info("-> Set worker to", uuid)
	}

	c.Logger.Info("Done scheduling")

	return nil
}
