package orchestrator

import (
	"github.com/gjy20/defense-agent/backend/internal/dag"
	"github.com/gjy20/defense-agent/backend/internal/types"
)

// RegisterAllScenes adds all 5 scene templates to the router
func (r *SceneRouter) RegisterAllScenes() {
	r.registerPentest()
	r.registerVulnResearch()
	r.registerReverseEng()
}

func (r *SceneRouter) registerPentest() {
	tmpl := &SceneTemplate{
		Scene:       types.ScenePenTest,
		Name:        "Penetration Test",
		Description: "Full penetration testing workflow: recon, scanning, exploitation, reporting",
		AgentOrder:  []types.AgentType{types.AgentResearcher, types.AgentDeveloper, types.AgentExecutor, types.AgentAuditor},
		BuildDAG: func(taskID string) *dag.DAG {
			d := dag.NewDAG("dag-"+taskID, taskID, types.ScenePenTest)

			recon := &dag.Node{
				ID: "researcher-"+taskID, AgentType: types.AgentResearcher,
				State: types.NodePending, MaxRetries: 2, Timeout: 120,
			}
			plan := &dag.Node{
				ID: "developer-"+taskID, AgentType: types.AgentDeveloper,
				State: types.NodePending, Dependencies: []string{"researcher-" + taskID},
				MaxRetries: 2, Timeout: 60,
			}
			exec := &dag.Node{
				ID: "executor-"+taskID, AgentType: types.AgentExecutor,
				State: types.NodePending, Dependencies: []string{"developer-" + taskID},
				MaxRetries: 3, Timeout: 300,
			}
			audit := &dag.Node{
				ID: "auditor-"+taskID, AgentType: types.AgentAuditor,
				State: types.NodePending, Dependencies: []string{"executor-" + taskID},
				MaxRetries: 1, Timeout: 60,
			}

			d.AddNode(recon)
			d.AddNode(plan)
			d.AddNode(exec)
			d.AddNode(audit)
			return d
		},
	}
	r.scenes[types.ScenePenTest] = tmpl
}

func (r *SceneRouter) registerVulnResearch() {
	tmpl := &SceneTemplate{
		Scene:       types.SceneVulnResearch,
		Name:        "Vulnerability Research",
		Description: "Targeted vulnerability discovery: CVE search, fuzzing, PoC development",
		AgentOrder:  []types.AgentType{types.AgentResearcher, types.AgentDeveloper, types.AgentExecutor, types.AgentMemorist},
		BuildDAG: func(taskID string) *dag.DAG {
			d := dag.NewDAG("dag-"+taskID, taskID, types.SceneVulnResearch)

			search := &dag.Node{
				ID: "researcher-"+taskID, AgentType: types.AgentResearcher,
				State: types.NodePending, MaxRetries: 2, Timeout: 120,
			}
			analyze := &dag.Node{
				ID: "developer-"+taskID, AgentType: types.AgentDeveloper,
				State: types.NodePending, Dependencies: []string{"researcher-" + taskID},
				MaxRetries: 2, Timeout: 90,
			}
			exploit := &dag.Node{
				ID: "executor-"+taskID, AgentType: types.AgentExecutor,
				State: types.NodePending, Dependencies: []string{"developer-" + taskID},
				MaxRetries: 3, Timeout: 300,
			}
			store := &dag.Node{
				ID: "memorist-"+taskID, AgentType: types.AgentMemorist,
				State: types.NodePending, Dependencies: []string{"executor-" + taskID},
				MaxRetries: 1, Timeout: 30,
			}

			d.AddNode(search)
			d.AddNode(analyze)
			d.AddNode(exploit)
			d.AddNode(store)
			return d
		},
	}
	r.scenes[types.SceneVulnResearch] = tmpl
}

func (r *SceneRouter) registerReverseEng() {
	tmpl := &SceneTemplate{
		Scene:       types.SceneReverseEng,
		Name:        "Reverse Engineering",
		Description: "Binary/malware analysis: static analysis, dynamic analysis, decompilation, IOC extraction",
		AgentOrder:  []types.AgentType{types.AgentPerceiver, types.AgentAnalyst, types.AgentExecutor, types.AgentMemorist, types.AgentAuditor},
		BuildDAG: func(taskID string) *dag.DAG {
			d := dag.NewDAG("dag-"+taskID, taskID, types.SceneReverseEng)

			static := &dag.Node{
				ID: "perceiver-static-"+taskID, AgentType: types.AgentPerceiver,
				State: types.NodePending, MaxRetries: 2, Timeout: 180,
			}
			dynamic := &dag.Node{
				ID: "perceiver-dynamic-"+taskID, AgentType: types.AgentPerceiver,
				State: types.NodePending, MaxRetries: 2, Timeout: 300,
			}
			analyze := &dag.Node{
				ID: "analyst-"+taskID, AgentType: types.AgentAnalyst,
				State: types.NodePending,
				Dependencies: []string{"perceiver-static-" + taskID, "perceiver-dynamic-" + taskID},
				MaxRetries: 2, Timeout: 120,
			}
			extract := &dag.Node{
				ID: "executor-"+taskID, AgentType: types.AgentExecutor,
				State: types.NodePending, Dependencies: []string{"analyst-" + taskID},
				MaxRetries: 2, Timeout: 180,
			}
			memorize := &dag.Node{
				ID: "memorist-"+taskID, AgentType: types.AgentMemorist,
				State: types.NodePending, Dependencies: []string{"executor-" + taskID},
				MaxRetries: 1, Timeout: 30,
			}

			d.AddNode(static)
			d.AddNode(dynamic)
			d.AddNode(analyze)
			d.AddNode(extract)
			d.AddNode(memorize)
			return d
		},
	}
	r.scenes[types.SceneReverseEng] = tmpl
}
