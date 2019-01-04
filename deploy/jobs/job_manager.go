package jobs

import (
	"fmt"
	"path/filepath"
	"strings"

	compilers "github.com/hyperledger/burrow/deploy/compile"
	pbpayload "github.com/hyperledger/burrow/txs/payload"

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
	case *def.Meta:
		for _, job := range job.Meta.Playbook.Jobs {
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

func doJobs(script *def.Playbook, args *def.DeployArgs, client *def.Client) error {
	for _, job := range script.Jobs {
		payload, err := job.Payload()
		if err != nil {
			return fmt.Errorf("could not get Job payload: %v", payload)
		}

		err = util.PreProcessFields(payload, args, script, client)
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
			job.Result, err = ProposalJob(job.Proposal, args, script, client)

		// Meta Job
		case *def.Meta:
			announce(job.Name, "Meta")
			metaPlaybook := job.Meta.Playbook
			if metaPlaybook.Account == "" {
				metaPlaybook.Account = script.Account
			}
			err = doJobs(metaPlaybook, args, client)

		// Governance
		case *def.UpdateAccount:
			announce(job.Name, "UpdateAccount")
			var tx *pbpayload.GovTx
			tx, job.Variables, err = FormulateUpdateAccountJob(job.UpdateAccount, script.Account, client)
			if err != nil {
				return err
			}
			err = UpdateAccountJob(job.UpdateAccount, script.Account, tx, client)

		// Util jobs
		case *def.Account:
			announce(job.Name, "Account")
			job.Result, err = SetAccountJob(job.Account, args, script)
		case *def.Set:
			announce(job.Name, "Set")
			job.Result, err = SetValJob(job.Set, args)

		// Transaction jobs
		case *def.Send:
			announce(job.Name, "Send")
			tx, err := FormulateSendJob(job.Send, script.Account, client)
			if err != nil {
				return err
			}
			job.Result, err = SendJob(job.Send, tx, script.Account, client)
		case *def.RegisterName:
			announce(job.Name, "RegisterName")
			txs, err := FormulateRegisterNameJob(job.RegisterName, args, script.Account, client)
			if err != nil {
				return err
			}
			job.Result, err = RegisterNameJob(job.RegisterName, args, script, txs, client)
		case *def.Permission:
			announce(job.Name, "Permission")
			tx, err := FormulatePermissionJob(job.Permission, script.Account, client)
			if err != nil {
				return err
			}
			job.Result, err = PermissionJob(job.Permission, script.Account, tx, client)

		// Contracts jobs
		case *def.Deploy:
			announce(job.Name, "Deploy")
			txs, contracts, ferr := FormulateDeployJob(job.Deploy, args, script, client, job.Intermediate)
			if ferr != nil {
				return ferr
			}
			job.Result, err = DeployJob(job.Deploy, args, script, client, txs, contracts)

		case *def.Call:
			announce(job.Name, "Call")
			CallTx, ferr := FormulateCallJob(job.Call, args, script, client)
			if ferr != nil {
				return ferr
			}
			job.Result, job.Variables, err = CallJob(job.Call, CallTx, args, script, client)
		case *def.Build:
			announce(job.Name, "Build")
			var resp *compilers.Response
			resp, err = getCompilerWork(job.Intermediate)
			if err != nil {
				return err
			}
			job.Result, err = BuildJob(job.Build, args.BinPath, resp)

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
			job.Result, job.Variables, err = QueryContractJob(job.QueryContract, args, script, client)
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

	return nil
}

func ExecutePlaybook(args *def.DeployArgs, script *def.Playbook, client *def.Client) error {
	// ADD DefaultAddr and DefaultSet to jobs array....
	// These work in reverse order and the addendums to the
	// the ordering from the loading process is lifo
	if len(args.DefaultSets) >= 1 {
		defaultSetJobs(args, script)
	}

	if args.Address != "" {
		defaultAddrJob(args, script)
	}

	err := args.Validate()
	if err != nil {
		return fmt.Errorf("error validating Burrow deploy file at %s: %v", args.YAMLPath, err)
	}

	// Ensure we have a queue large enough so that we don't have to wait for more work to be queued
	jobs := make(chan *intermediateJob, args.Jobs*2)
	defer close(jobs)

	for i := 0; i < args.Jobs; i++ {
		go intermediateJobRunner(jobs)
	}

	for _, job := range script.Jobs {
		queueCompilerWork(job, jobs)
	}

	err = doJobs(script, args, client)
	if err != nil {
		return err
	}

	postProcess(args, script)
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

func defaultAddrJob(do *def.DeployArgs, deployScript *def.Playbook) {
	oldJobs := deployScript.Jobs

	newJob := &def.Job{
		Name: "defaultAddr",
		Account: &def.Account{
			Address: do.Address,
		},
	}

	deployScript.Jobs = append([]*def.Job{newJob}, oldJobs...)
}

func defaultSetJobs(do *def.DeployArgs, deployScript *def.Playbook) {
	oldJobs := deployScript.Jobs

	newJobs := []*def.Job{}

	for _, setr := range do.DefaultSets {
		blowdUp := strings.Split(setr, "=")
		if len(blowdUp) == 2 && blowdUp[0] != "" {
			newJobs = append(newJobs, &def.Job{
				Name: blowdUp[0],
				Set: &def.Set{
					Value: blowdUp[1],
				},
			})
		}
	}

	deployScript.Jobs = append(newJobs, oldJobs...)
}

func postProcess(do *def.DeployArgs, deployScript *def.Playbook) error {
	// Formulate the results map
	results := make(map[string]interface{})
	for _, job := range deployScript.Jobs {
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
