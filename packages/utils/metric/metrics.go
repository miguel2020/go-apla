// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package metric

import (
	"strconv"
	"time"

	"github.com/AplaProject/go-apla/packages/consts"
	"github.com/AplaProject/go-apla/packages/model"

	log "github.com/sirupsen/logrus"
)

const (
	metricEcosystemPages   = "ecosystem_pages"
	metricEcosystemMembers = "ecosystem_members"
	metricEcosystemTx      = "ecosystem_tx"
)

// CollectMetricDataForEcosystemTables returns metrics for some tables of ecosystems
func CollectMetricDataForEcosystemTables(timeBlock int64) (metricValues []*Value, err error) {
	stateIDs, _, err := model.GetAllSystemStatesIDs()
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("get all system states ids")
		return nil, err
	}

	now := time.Unix(timeBlock, 0)
	unixDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local).Unix()

	for _, stateID := range stateIDs {
		var pagesCount, membersCount int64

		tablePrefix := strconv.FormatInt(stateID, 10)

		p := &model.Page{}
		p.SetTablePrefix(tablePrefix)
		if pagesCount, err = p.Count(); err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("get count of pages")
			return nil, err
		}
		metricValues = append(metricValues, &Value{
			Time:   unixDate,
			Metric: metricEcosystemPages,
			Key:    tablePrefix,
			Value:  pagesCount,
		})

		m := &model.Member{}
		m.SetTablePrefix(tablePrefix)
		if membersCount, err = m.Count(); err != nil {
			log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("get count of members")
			return nil, err
		}
		metricValues = append(metricValues, &Value{
			Time:   unixDate,
			Metric: metricEcosystemMembers,
			Key:    tablePrefix,
			Value:  membersCount,
		})
	}

	return metricValues, nil
}

// CollectMetricDataForEcosystemTx returns metrics for transactions of ecosystems
func CollectMetricDataForEcosystemTx(timeBlock int64) (metricValues []*Value, err error) {
	ecosystemTx, err := model.GetEcosystemTxPerDay(timeBlock)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "type": consts.DBError}).Error("get ecosystem transactions by period")
		return nil, err
	}
	for _, item := range ecosystemTx {
		if len(item.Ecosystem) == 0 {
			continue
		}

		metricValues = append(metricValues, &Value{
			Time:   item.UnixTime,
			Metric: metricEcosystemTx,
			Key:    item.Ecosystem,
			Value:  item.Count,
		})
	}

	return metricValues, nil
}
