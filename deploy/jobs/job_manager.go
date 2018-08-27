package jobs

import (
	"fmt"
	"path/filepath"
	"strings"

	compilers "github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/util"
	log "github.com/sirupsen/logrus"
)

type trackJob struct {
	payload      def.Payload
	job          *def.Job
	contractName string
	compilerResp *compilers.Response
	err          error
	done         chan struct{}
}

func concurrentJobRunner(jobs chan *trackJob) {
	for {
		track, ok := <-jobs
		if !ok {
			break
		}
		(*track).compilerResp, (*track).err = compilers.Compile(track.contractName, false, nil)
		close(track.done)
	}
}

func DoJobs(do *def.Packages) error {
	jobs := make(chan *trackJob, do.Jobs*2)
	defer close(jobs)
	return RunJobs(do, jobs)
}

func RunJobs(do *def.Packages, jobs chan *trackJob) error {

	// Dial the chain if needed
	err, needed := burrowConnectionNeeded(do)
	if err != nil {
		return err
	}

	if needed {
		err = do.Dial()
		if err != nil {
			return err
		}
	}

	// ADD DefaultAddr and DefaultSet to jobs array....
	// These work in reverse order and the addendums to the
	// the ordering from the loading process is lifo
	if len(do.DefaultSets) >= 1 {
		defaultSetJobs(do)
	}

	if do.Address != "" {
		defaultAddrJob(do)
	}

	err = do.Validate()
	if err != nil {
		return fmt.Errorf("error validating Burrow deploy file at %s: %v", do.YAMLPath, err)
	}

	intermediateJobs := make([]*trackJob, 0, len(do.Package.Jobs))

	for i := 0; i < do.Jobs; i++ {
		go concurrentJobRunner(jobs)
	}

	for _, job := range do.Package.Jobs {
		payload, err := job.Payload()
		if err != nil {
			return fmt.Errorf("could not get Job payload: %v", payload)
		}

		track := trackJob{job: job, payload: payload}
		intermediateJobs = append(intermediateJobs, &track)

		// Do compilation first
		switch payload.(type) {
		case *def.Build:
			track.done = make(chan struct{})
			track.contractName = job.Build.Contract
			jobs <- &track
		case *def.Deploy:
			if filepath.Ext(job.Deploy.Contract) == ".sol" {
				track.done = make(chan struct{})
				track.contractName = job.Deploy.Contract
				jobs <- &track
			}
		}
	}

	for _, m := range intermediateJobs {
		job := m.job

		err = util.PreProcessFields(m.payload, do)
		if err != nil {
			return err
		}
		// Revalidate with possible replacements
		err = m.payload.Validate()
		if err != nil {
			return fmt.Errorf("error validating job %s after pre-processing variables: %v", job.Name, err)
		}

		if m.done != nil {
			<-m.done

			if m.err != nil {
				return m.err
			}
		}

		switch m.payload.(type) {
		// Meta Job
		case *def.Meta:
			announce(job.Name, "Meta")
			do.CurrentOutput = fmt.Sprintf("%s.output.json", job.Name)
			job.Result, err = MetaJob(job.Meta, do, jobs)

		// Governance
		case *def.UpdateAccount:
			announce(job.Name, "UpdateAccount")
			job.Result, job.Variables, err = UpdateAccountJob(job.UpdateAccount, do)

		// Util jobs
		case *def.Account:
			announce(job.Name, "Account")
			job.Result, err = SetAccountJob(job.Account, do)
		case *def.Set:
			announce(job.Name, "Set")
			job.Result, err = SetValJob(job.Set, do)

		// Transaction jobs
		case *def.Send:
			announce(job.Name, "Sent")
			job.Result, err = SendJob(job.Send, do)
		case *def.RegisterName:
			announce(job.Name, "RegisterName")
			job.Result, err = RegisterNameJob(job.RegisterName, do)
		case *def.Permission:
			announce(job.Name, "Permission")
			job.Result, err = PermissionJob(job.Permission, do)

		// Contracts jobs
		case *def.Deploy:
			announce(job.Name, "Deploy")
			job.Result, err = DeployJob(job.Deploy, do, m.compilerResp)
		case *def.Call:
			announce(job.Name, "Call")
			job.Result, job.Variables, err = CallJob(job.Call, do)
		case *def.Build:
			announce(job.Name, "Build")
			job.Result, err = BuildJob(job.Build, do, m.compilerResp)

		// State jobs
		case *def.RestoreState:
			announce(job.Name, "RestoreState")
			job.Result, err = RestoreStateJob(job.RestoreState, do)
		case *def.DumpState:
			announce(job.Name, "DumpState")
			job.Result, err = DumpStateJob(job.DumpState, do)

		// Test jobs
		case *def.QueryAccount:
			announce(job.Name, "QueryAccount")
			job.Result, err = QueryAccountJob(job.QueryAccount, do)
		case *def.QueryContract:
			announce(job.Name, "QueryContract")
			job.Result, job.Variables, err = QueryContractJob(job.QueryContract, do)
		case *def.QueryName:
			announce(job.Name, "QueryName")
			job.Result, err = QueryNameJob(job.QueryName, do)
		case *def.QueryVals:
			announce(job.Name, "QueryVals")
			job.Result, err = QueryValsJob(job.QueryVals, do)
		case *def.Assert:
			announce(job.Name, "Assert")
			job.Result, err = AssertJob(job.Assert, do)

		default:
			log.Error("")
			return fmt.Errorf("the Job specified in deploy.yaml and parsed as '%v' is not recognised as a valid job",
				job)
		}

		if len(job.Variables) != 0 {
			for _, theJob := range job.Variables {
				log.WithField("=>", fmt.Sprintf("%s,%s", theJob.Name, theJob.Value)).Info("Job Vars")
			}
		}

		if err != nil {
			return err
		}
	}
	postProcess(do)
	return nil
}

func announce(job, typ string) {
	log.Warn("*****Executing Job*****\n")
	log.WithField("=>", job).Warn("Job Name")
	log.WithField("=>", typ).Info("Type")
	log.Warn("\n")
}

func defaultAddrJob(do *def.Packages) {
	oldJobs := do.Package.Jobs

	newJob := &def.Job{
		Name: "defaultAddr",
		Account: &def.Account{
			Address: do.Address,
		},
	}

	do.Package.Jobs = append([]*def.Job{newJob}, oldJobs...)
}

func defaultSetJobs(do *def.Packages) {
	oldJobs := do.Package.Jobs

	newJobs := []*def.Job{}

	for _, setr := range do.DefaultSets {
		blowdUp := strings.Split(setr, "=")
		if blowdUp[0] != "" {
			newJobs = append(newJobs, &def.Job{
				Name: blowdUp[0],
				Set: &def.Set{
					Value: blowdUp[1],
				},
			})
		}
	}

	do.Package.Jobs = append(newJobs, oldJobs...)
}

func postProcess(do *def.Packages) error {
	// Formulate the results map
	results := make(map[string]interface{})
	for _, job := range do.Package.Jobs {
		results[job.Name] = job.Result
	}

	// check do.YAMLPath and do.DefaultOutput
	var yaml string
	yamlName := strings.LastIndexByte(do.YAMLPath, '.')
	if yamlName >= 0 {
		yaml = do.YAMLPath[:yamlName]
	} else {
		return fmt.Errorf("invalid jobs file path (%s)", do.YAMLPath)
	}

	// if do.YAMLPath is not default and do.DefaultOutput is default, over-ride do.DefaultOutput
	if yaml != "deploy" && do.DefaultOutput == "deploy.output.json" {
		do.DefaultOutput = fmt.Sprintf("%s.output.json", yaml)
	}

	// if CurrentOutput set, we're in a meta job
	if do.CurrentOutput != "" {
		log.Warn(fmt.Sprintf("Writing meta output of [%s] to current directory", do.CurrentOutput))
		return WriteJobResultJSON(results, do.CurrentOutput)
	}

	// Write the output
	log.Warn(fmt.Sprintf("Writing [%s] to current directory", do.DefaultOutput))
	return WriteJobResultJSON(results, do.DefaultOutput)
}

func burrowConnectionNeeded(do *def.Packages) (error, bool) {
	// Dial the chain if needed
	for _, job := range do.Package.Jobs {
		payload, err := job.Payload()
		if err != nil {
			return fmt.Errorf("could not get Job payload: %v", payload), false
		}
		switch payload.(type) {
		case *def.Meta:
			// A meta jobs will call runJobs again, so it does not need a connection for itself
			continue
		case *def.Build:
			continue
		case *def.Set:
			continue
		default:
			return nil, true
		}
	}

	return nil, false
}
