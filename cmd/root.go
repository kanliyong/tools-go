package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func filterClusters(env string) []*Cluster {
	switch env {
	case "all":
		return ClusterArray
	case "test":
		return ClusterArray[:2]
	case "stage":
		return ClusterArray[2:4]
	case "product":
		return ClusterArray[4:]
	}
	return ClusterArray
}

func FindDeploy(srv string, f bool, show bool, env string) (result string, err error) {
	for _, cluster := range filterClusters(env) {
		args := []string{
			"-n",
			cluster.env,
			"--kubeconfig",
			"kubeconfig/" + cluster.kubeconfig,
			"get",
			"deploy",
		}
		output, err := exec.Command("kubectl", args...).Output()
		if err != nil {
			panic(err)
		}
		re := regexp.MustCompile(srv + `\S*`)
		deployName := re.FindString(string(output))
		if deployName != "" {
			log.Printf("%s", deployName)
			if show == true {
			} else {
				ScaleZero(cluster, deployName, f)
			}
		}

	}

	return result, nil
}

func ScaleZero(cluster *Cluster, svc string, f bool) {
	args := []string{
		"scale",
		"-n",
		cluster.env,
		"--kubeconfig",
		"kubeconfig/" + cluster.kubeconfig,
		"deployments/" + svc,
		"--replicas=0",
	}

	if f == true {
		doScale(args)

	} else {
		var input string
		fmt.Printf("Are you sure sacle to 0 depley=%s,cluser=%v ? Y/N \n", svc, cluster)
		fmt.Scanf("%s", &input)
		if input == "Y" {
			doScale(args)
		}
	}
}

func doScale(args []string) {
	output, err := exec.Command("kubectl", args...).Output()
	if err != nil {
		panic(err)
	}
	fmt.Println(string(output))
}
func doDeleteDocument(coll *mongo.Collection, result *bson.M) {
	deleteReuslt, err := coll.DeleteOne(context.TODO(), bson.D{{"_id", (*result)["_id"]}})
	if err != nil {
		panic(err)
	}
	fmt.Println("delete documents count =", deleteReuslt.DeletedCount)
}

func RmMongo(svc string, f bool, show bool, env string) {
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://192.168.0.18:27010/data-mgr"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	coll := client.Database("data-mgr").Collection("devex_request")

	filter := bson.D{{
		"projectid",
		bson.D{{"$regex", "^" + svc}},
	}}
	cur, err := coll.Find(context.TODO(), filter)
	if err != nil {
		panic(err)
	}
	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var input string
		var result bson.M
		err := cur.Decode(&result)
		if err != nil {
			log.Panic(err)
		}

		deployEnv := result["deployenv"].(string)
		if env == "all" {

		} else if env == "test" && !strings.Contains(deployEnv, "test") {
			continue
		} else if env == "stage" && !strings.Contains(deployEnv, "proenv") {
			continue
		} else if env == "product" && strings.Contains(deployEnv, "release") {
			continue
		}

		r, _ := json.MarshalIndent(result, "", "\t")
		log.Println(string(r))

		if show == true {

		} else {
			if f == true {
				doDeleteDocument(coll, &result)
			} else {
				fmt.Printf("Are you sure delete document %s; Y/N \n", result["_id"])
				fmt.Scanf("%s", &input)
				if input == "Y" {
					doDeleteDocument(coll, &result)
				}
			}
		}
	}
}
