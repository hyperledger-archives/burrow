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

type intermediateJob struct {
	contractName string
	compilerResp *compilers.Response
	err          error
	done         chan struct{}
}

func intermediateJobRunner(jobs chan *intermediateJob) {
	for {
		intermediate, ok := <-jobs
		if !ok {
			break
		}
		resp, err := compilers.Compile(intermediate.contractName, false, nil)
		(*intermediate).compilerResp = resp
		(*intermediate).err = err
		close(intermediate.done)
	}
}

func queueCompilerWork(job *def.Job, jobs chan *intermediateJob) error {
	payload, err := job.Payload()
	if err != nil {
		return fmt.Errorf("could not get Job payload: %v", payload)
	}

	// Do compilation first
	switch payload.(type) {
	case *def.Build:
		intermediate := intermediateJob{done: make(chan struct{}), contractName: job.Build.Contract}
		job.Intermediate = &intermediate
		jobs <- &intermediate
	case *def.Deploy:
		if filepath.Ext(job.Deploy.Contract) == ".sol" {
			intermediate := intermediateJob{done: make(chan struct{}), contractName: job.Deploy.Contract}
			job.Intermediate = &intermediate
			jobs <- &intermediate
		}
	case *def.Proposal:
		for _, job := range job.Proposal.Jobs {
			err = queueCompilerWork(job, jobs)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getCompilerWork(intermediate interface{}) (*compilers.Response, error) {
	if intermediate, ok := intermediate.(*intermediateJob); ok {
		<-intermediate.done

		return intermediate.compilerResp, intermediate.err
	}

	return nil, fmt.Errorf("internal error: no compiler work queued")
}

func DoJobs(do *def.DeployArgs, client *def.Client) error {
	// ADD DefaultAddr and DefaultSet to jobs array....
	// These work in reverse order and the addendums to the
	// the ordering from the loading process is lifo
	if len(do.DefaultSets) >= 1 {
		defaultSetJobs(do)
	}

	if do.Address != "" {
		defaultAddrJob(do)
	}

	err := do.Validate()
	if err != nil {
		return fmt.Errorf("error validating Burrow deploy file at %s: %v", do.YAMLPath, err)
	}

	// Ensure we have a queue large enough so that we don't have to wait for more work to be queued
	jobs := make(chan *intermediateJob, do.Jobs*2)
	defer close(jobs)

	for i := 0; i < do.Jobs; i++ {
		go intermediateJobRunner(jobs)
	}

	for _, job := range do.Package.Jobs {
		queueCompilerWork(job, jobs)
	}

	for _, job := range do.Package.Jobs {
		payload, err := job.Payload()
		if err != nil {
			return fmt.Errorf("could not get Job payload: %v", payload)
		}

		err = util.PreProcessFields(payload, do, client)
		if err != nil {
			return err
		}
		// Revalidate with possible replacements
		err = payload.Validate()
		if err != nil {
			return fmt.Errorf("error validating job %s after pre-processing variables: %v", job.Name, err)
		}

		switch payload.(type) {
		case *def.Proposal:
			announce(job.Name, "Proposal")
			job.Result, err = ProposalJob(job.Proposal, do, client)

		// Meta Job
		case *def.Meta:
			announce(job.Name, "Meta")
			do.CurrentOutput = fmt.Sprintf("%s.output.json", job.Name)
			job.Result, err = MetaJob(job.Meta, do, client)

		// Governance
		case *def.UpdateAccount:
			announce(job.Name, "UpdateAccount")
			job.Result, job.Variables, err = UpdateAccountJob(job.UpdateAccount, do.Package.Account, client)

		// Util jobs
		case *def.Account:
			announce(job.Name, "Account")
			job.Result, err = SetAccountJob(job.Account, do)
		case *def.Set:
			announce(job.Name, "Set")
			job.Result, err = SetValJob(job.Set, do)

		// Transaction jobs
		case *def.Send:
			announce(job.Name, "Send")
			tx, err := FormulateSendJob(job.Send, do.Package.Account, client)
			if err != nil {
				return err
			}
			job.Result, err = SendJob(job.Send, tx, do.Package.Account, client)
		case *def.RegisterName:
			announce(job.Name, "RegisterName")
			txs, err := FormulateRegisterNameJob(job.RegisterName, do, client)
			if err != nil {
				return err
			}
			job.Result, err = RegisterNameJob(job.RegisterName, do, txs, client)
		case *def.Permission:
			announce(job.Name, "Permission")
			tx, err := FormulatePermissionJob(job.Permission, do.Package.Account, client)
			if err != nil {
				return err
			}
			job.Result, err = PermissionJob(job.Permission, do.Package.Account, tx, client)

		// Contracts jobs
		case *def.Deploy:
			announce(job.Name, "Deploy")
			txs, contracts, ferr := FormulateDeployJob(job.Deploy, do, client, job.Intermediate)
			if ferr != nil {
				return ferr
			}
			job.Result, err = DeployJob(job.Deploy, do, client, txs, contracts)

		case *def.Call:
			announce(job.Name, "Call")
			CallTx, ferr := FormulateCallJob(job.Call, do, client)
			if ferr != nil {
				return ferr
			}
			job.Result, job.Variables, err = CallJob(job.Call, CallTx, do, client)
		case *def.Build:
			announce(job.Name, "Build")
			var resp *compilers.Response
			resp, err = getCompilerWork(job.Intermediate)
			if err != nil {
				return err
			}
			job.Result, err = BuildJob(job.Build, do.BinPath, resp)

		// State jobs
		case *def.RestoreState:
			announce(job.Name, "RestoreState")
			job.Result, err = RestoreStateJob(job.RestoreState)
		case *def.DumpState:
			announce(job.Name, "DumpState")
			job.Result, err = DumpStateJob(job.DumpState)

		// Test jobs
		case *def.QueryAccount:
			announce(job.Name, "QueryAccount")
			job.Result, err = QueryAccountJob(job.QueryAccount, client)
		case *def.QueryContract:
			announce(job.Name, "QueryContract")
			job.Result, job.Variables, err = QueryContractJob(job.QueryContract, do, client)
		case *def.QueryName:
			announce(job.Name, "QueryName")
			job.Result, err = QueryNameJob(job.QueryName, client)
		case *def.QueryVals:
			announce(job.Name, "QueryVals")
			job.Result, err = QueryValsJob(job.QueryVals, client)
		case *def.Assert:
			announce(job.Name, "Assert")
			job.Result, err = AssertJob(job.Assert)

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

func announceProposalJob(job, typ string) {
	log.Warn("*****Capturing Proposal Job*****\n")
	log.WithField("=>", job).Warn("Job Name")
	log.WithField("=>", typ).Info("Type")
	log.Warn("\n")
}

func defaultAddrJob(do *def.DeployArgs) {
	oldJobs := do.Package.Jobs

	newJob := &def.Job{
		Name: "defaultAddr",
		Account: &def.Account{
			Address: do.Address,
		},
	}

	do.Package.Jobs = append([]*def.Job{newJob}, oldJobs...)
}

func defaultSetJobs(do *def.DeployArgs) {
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

func postProcess(do *def.DeployArgs) error {
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
