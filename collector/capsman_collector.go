package collector

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type capsmanCollector struct {
	capsmanClientCountDesc *prometheus.Desc
}

func (c *capsmanCollector) init() {
	const prefix = "capsman"

	labelNames := []string{"name", "address", "capname", "ssid"}
	c.capsmanClientCountDesc = description(prefix, "clients_active_count", "number of connected clients per CAP", labelNames)
}

func newCAPSManCollector() routerOSCollector {
	c := &capsmanCollector{}
	c.init()
	return c
}

func (c *capsmanCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.capsmanClientCountDesc
}

func (c *capsmanCollector) collect(ctx *collectorContext) error {
	mappings, err := c.fetchSSIDMappings(ctx)
	if err != nil {
		return err
	}

	err = c.colllectForCAP(ctx, mappings)
	if err != nil {
		return err
	}

	return nil
}

func (c *capsmanCollector) fetchSSIDMappings(ctx *collectorContext) (map[string]string, error) {
        reply, err := ctx.client.Run("/caps-man/configuration/print", "=.proplist=name,ssid")
        if err != nil {
                log.WithFields(log.Fields{
                        "device": ctx.device.Name,
                        "error":  err,
                }).Error("error fetching SSID mappings")
                return nil, err
        }

        mappings := make(map[string]string)
        for _, re := range reply.Re {
                mappings[re.Map["name"]] = re.Map["ssid"]
        }

	log.Info(mappings)

        return mappings, nil
}


func (c *capsmanCollector) colllectForCAP(ctx *collectorContext, mappings map[string]string) error {
	reply, err := ctx.client.Run("/caps-man/interface/print", "=.proplist=name,configuration,current-registered-clients")
	if err != nil {
		log.WithFields(log.Fields{
			"device":      ctx.device.Name,
			"error":       err,
		}).Error("error fetching CAP client counts")
		return err
	}

        for _, re := range reply.Re {
		v, err := strconv.ParseFloat(re.Map["current-registered-clients"], 32)

		if err != nil {
			log.WithFields(log.Fields{
				"device":      ctx.device.Name,
				"error":       err,
			}).Error("error parsing CAP client counts")
			return err
		}

		name := re.Map["name"]
		ssid := mappings[re.Map["configuration"]]

		ctx.ch <- prometheus.MustNewConstMetric(c.capsmanClientCountDesc, prometheus.GaugeValue, v, ctx.device.Name, ctx.device.Address, name, ssid)
	}
	return nil
}
