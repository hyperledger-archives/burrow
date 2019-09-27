package jobs

import (
	"fmt"
	"path/filepath"
	"strings"

	compilers "github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/util"
	"github.com/hyperledger/burrow/logging"
	pbpayload "github.com/hyperledger/burrow/txs/payload"
)

const (
	// How many concurrent
	concurrentSolc = 2
	// Ensure we have a queue large enough so that we don't have to wait for more work to be queued
	concurrentSolcWorkQueue = 4
)

type solidityCompilerWork struct {
	contractName string
	workDir      string
}

type compilerJob struct {
	work         solidityCompilerWork
	compilerResp *compilers.Response
	err          error
	done         chan struct{}
}

func solcRunner(jobs chan *compilerJob, logger *logging.Logger) {
	for {
		job, ok := <-jobs
		if !ok {
			break
		}
		resp, err := compilers.EVM(job.work.contractName, false, job.work.workDir, nil, logger)
		(*job).compilerResp = resp
		(*job).err = err
		close(job.done)
	}
}

func solangRunner(jobs chan *compilerJob, logger *logging.Logger) {
	for {
		job, ok := <-jobs
		if !ok {
			return
		}
		resp, err := compilers.WASM(job.work.contractName, job.work.workDir, logger)
		(*job).compilerResp = resp
		(*job).err = err
		close(job.done)
	}
}

func queueCompilerWork(job *def.Job, playbook *def.Playbook, jobs chan *compilerJob) error {
	payload, err := job.Payload()
	if err != nil {
		return fmt.Errorf("could not get Job payload: %v", payload)
	}

	// Do compilation first
	switch payload.(type) {
	case *def.Build:
		intermediate := compilerJob{
			done: make(chan struct{}),
			work: solidityCompilerWork{
				contractName: job.Build.Contract,
				workDir:      playbook.Path,
			},
		}
		job.Intermediate = &intermediate
		jobs <- &intermediate
	case *def.Deploy:
		if filepath.Ext(job.Deploy.Contract) == ".sol" {
			intermediate := compilerJob{
				done: make(chan struct{}),
				work: solidityCompilerWork{
					contractName: job.Deploy.Contract,
					workDir:      playbook.Path,
				},
			}
			job.Intermediate = &intermediate
			jobs <- &intermediate
		}
	case *def.Proposal:
		for _, job := range job.Proposal.Jobs {
			err = queueCompilerWork(job, playbook, jobs)
			if err != nil {
				return err
			}
		}
	case *def.Meta:
		for _, job := range job.Meta.Playbook.Jobs {
			err = queueCompilerWork(job, playbook, jobs)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func getCompilerWork(intermediate interface{}) (*compilers.Response, error) {
	if intermediate, ok := intermediate.(*compilerJob); ok {
		<-intermediate.done

		return intermediate.compilerResp, intermediate.err
	}

	return nil, fmt.Errorf("internal error: no compiler work queued")
}

func doJobs(playbook *def.Playbook, args *def.DeployArgs, client *def.Client, logger *logging.Logger) error {
	for _, job := range playbook.Jobs {
		payload, err := job.Payload()
		if err != nil {
			return fmt.Errorf("could not get Job payload: %v", payload)
		}

		err = util.PreProcessFields(payload, args, playbook, client, logger)
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
			announce(job.Name, "Proposal", logger)
			job.Result, err = ProposalJob(job.Proposal, args, playbook, client, logger)

		// Meta Job
		case *def.Meta:
			announce(job.Name, "Meta", logger)
			metaPlaybook := job.Meta.Playbook
			if metaPlaybook.Account == "" {
				metaPlaybook.Account = playbook.Account
			}
			err = doJobs(metaPlaybook, args, client, logger)

		// Governance
		case *def.UpdateAccount:
			announce(job.Name, "UpdateAccount", logger)
			var tx *pbpayload.GovTx
			tx, job.Variables, err = FormulateUpdateAccountJob(job.UpdateAccount, playbook.Account, client, logger)
			if err != nil {
				return err
			}
			err = UpdateAccountJob(job.UpdateAccount, playbook.Account, tx, client, logger)

		// Util jobs
		case *def.Account:
			announce(job.Name, "Account", logger)
			job.Result, err = SetAccountJob(job.Account, args, playbook, logger)
		case *def.Set:
			announce(job.Name, "Set", logger)
			job.Result, err = SetValJob(job.Set, args, logger)

		// Transaction jobs
		case *def.Send:
			announce(job.Name, "Send", logger)
			tx, err := FormulateSendJob(job.Send, playbook.Account, client, logger)
			if err != nil {
				return err
			}
			job.Result, err = SendJob(job.Send, tx, playbook.Account, client, logger)
			if err != nil {
				return err
			}
		case *def.Bond:
			announce(job.Name, "Bond", logger)
			tx, err := FormulateBondJob(job.Bond, playbook.Account, client, logger)
			if err != nil {
				return err
			}
			job.Result, err = BondJob(job.Bond, tx, playbook.Account, client, logger)
			if err != nil {
				return err
			}
		case *def.Unbond:
			announce(job.Name, "Unbond", logger)
			tx, err := FormulateUnbondJob(job.Unbond, playbook.Account, client, logger)
			if err != nil {
				return err
			}
			job.Result, err = UnbondJob(job.Unbond, tx, playbook.Account, client, logger)
			if err != nil {
				return err
			}
		case *def.RegisterName:
			announce(job.Name, "RegisterName", logger)
			txs, err := FormulateRegisterNameJob(job.RegisterName, args, playbook, client, logger)
			if err != nil {
				return err
			}
			job.Result, err = RegisterNameJob(job.RegisterName, args, playbook, txs, client, logger)
			if err != nil {
				return err
			}
		case *def.Permission:
			announce(job.Name, "Permission", logger)
			tx, err := FormulatePermissionJob(job.Permission, playbook.Account, client, logger)
			if err != nil {
				return err
			}
			job.Result, err = PermissionJob(job.Permission, playbook.Account, tx, client, logger)
			if err != nil {
				return err
			}
		case *def.Identify:
			announce(job.Name, "Identify", logger)
			tx, err := FormulateIdentifyJob(job.Identify, playbook.Account, client, logger)
			if err != nil {
				return err
			}
			job.Result, err = IdentifyJob(job.Identify, tx, playbook.Account, client, logger)
			if err != nil {
				return err
			}

		// Contracts jobs
		case *def.Deploy:
			announce(job.Name, "Deploy", logger)
			txs, contracts, ferr := FormulateDeployJob(job.Deploy, args, playbook, client, job.Intermediate, logger)
			if ferr != nil {
				return ferr
			}
			job.Result, err = DeployJob(job.Deploy, args, playbook, client, txs, contracts, logger)

		case *def.Call:
			announce(job.Name, "Call", logger)
			CallTx, ferr := FormulateCallJob(job.Call, args, playbook, client, logger)
			if ferr != nil {
				return ferr
			}
			job.Result, job.Variables, err = CallJob(job.Call, CallTx, args, playbook, client, logger)
		case *def.Build:
			announce(job.Name, "Build", logger)
			var resp *compilers.Response
			resp, err = getCompilerWork(job.Intermediate)
			if err != nil {
				return err
			}
			job.Result, err = BuildJob(job.Build, playbook, resp, logger)

		// State jobs
		case *def.RestoreState:
			announce(job.Name, "RestoreState", logger)
			job.Result, err = RestoreStateJob(job.RestoreState)
		case *def.DumpState:
			announce(job.Name, "DumpState", logger)
			job.Result, err = DumpStateJob(job.DumpState)

		// Test jobs
		case *def.QueryAccount:
			announce(job.Name, "QueryAccount", logger)
			job.Result, err = QueryAccountJob(job.QueryAccount, client, logger)
		case *def.QueryContract:
			announce(job.Name, "QueryContract", logger)
			job.Result, job.Variables, err = QueryContractJob(job.QueryContract, args, playbook, client, logger)
		case *def.QueryName:
			announce(job.Name, "QueryName", logger)
			job.Result, err = QueryNameJob(job.QueryName, client, logger)
		case *def.QueryVals:
			announce(job.Name, "QueryVals", logger)
			job.Result, err = QueryValsJob(job.QueryVals, client, logger)
		case *def.Assert:
			announce(job.Name, "Assert", logger)
			job.Result, err = AssertJob(job.Assert, logger)

		default:
			logger.InfoMsg("Error")
			return fmt.Errorf("the Job specified in deploy.yaml and parsed as '%v' is not recognised as a valid job",
				job)
		}

		if len(job.Variables) != 0 {
			for _, theJob := range job.Variables {
				logger.InfoMsg("Job Vars", "name", theJob.Name, "value", theJob.Value)
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func ExecutePlaybook(args *def.DeployArgs, playbook *def.Playbook, client *def.Client, logger *logging.Logger) error {
	// ADD DefaultAddr and DefaultSet to jobs array....
	// These work in reverse order and the addendums to the
	// the ordering from the loading process is lifo
	if len(args.DefaultSets) >= 1 {
		defaultSetJobs(args, playbook)
	}

	if args.Address != "" {
		defaultAddrJob(args, playbook)
	}

	err := args.Validate()
	if err != nil {
		return fmt.Errorf("error validating Burrow deploy file at %s: %v", playbook.Filename, err)
	}

	jobs := make(chan *compilerJob, concurrentSolcWorkQueue)
	defer close(jobs)

	for i := 0; i < concurrentSolc; i++ {
		if args.Wasm {
			go solangRunner(jobs, logger)
		} else {
			go solcRunner(jobs, logger)
		}
	}

	for _, job := range playbook.Jobs {
		queueCompilerWork(job, playbook, jobs)
	}

	err = doJobs(playbook, args, client, logger)
	if err != nil {
		return err
	}

	postProcess(args, playbook, logger)
	return nil
}

func announce(job, typ string, logger *logging.Logger) {
	logger.InfoMsg("*****Executing Job*****", "Job Name", job, "Type", typ)
}

func announceProposalJob(job, typ string, logger *logging.Logger) {
	logger.InfoMsg("*****Capturing Proposal Job*****", "Job Name", job, "Type", typ)
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

func postProcess(args *def.DeployArgs, playbook *def.Playbook, logger *logging.Logger) error {
	// Formulate the results map
	results := make(map[string]interface{})
	for _, job := range playbook.Jobs {
		results[job.Name] = job.Result
	}

	// check do.YAMLPath and do.DefaultOutput
	var yaml string
	yamlName := strings.LastIndexByte(playbook.Filename, '.')
	if yamlName >= 0 {
		yaml = playbook.Filename[:yamlName]
	} else {
		return fmt.Errorf("invalid jobs file path (%s)", playbook.Filename)
	}

	// if do.YAMLPath is not default and do.DefaultOutput is default, over-ride do.DefaultOutput
	if yaml != "deploy" && args.DefaultOutput == def.DefaultOutputFile {
		args.DefaultOutput = fmt.Sprintf("%s.output.json", yaml)
	}

	// if CurrentOutput set, we're in a meta job
	if args.CurrentOutput != "" {
		logger.InfoMsg("Writing meta output to current directory", "output", args.CurrentOutput)
		return WriteJobResultJSON(results, args.CurrentOutput)
	}

	// Write the output
	logger.InfoMsg("Writing to current directory", "output", args.DefaultOutput)
	return WriteJobResultJSON(results, args.DefaultOutput)
}
