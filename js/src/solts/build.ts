import { promises as fs } from "fs";
import * as path from "path";
import * as solcv5 from "solc_v5";
import * as solcv8 from "solc_v8";
import {
  decodeOutput,
  encodeInput,
  importLocalResolver,
  inputDescriptionFromFiles,
  Solidity
} from "../contracts/compile";
import { Compiled, newFile, printNodes, tokenizeLinks } from "./api";

const solcCompilers = {
  v5: solcv5,
  v8: solcv8
} as const;

export const defaultBuildOptions = {
  solcVersion: "v5" as keyof typeof solcCompilers,
  burrowImportPath: (sourceFile: string) => "@hyperledger/burrow" as string,
  binPath: "bin" as string,
  abiExt: ".abi" as string,
  // Used to resolve layout in bin folder - defaults to srcPath if is passed or process.cwd() otherwise
  basePath: undefined as undefined | string,
  failOnWarnings: false as boolean
} as const;

export type BuildOptions = typeof defaultBuildOptions;

/**
 * This is our Solidity -> Typescript code generation function, it:
 *  - Compiles Solidity source
 *  - Generates typescript code wrapping the Solidity contracts and functions that calls Burrow
 *  - Generates typescript code to deploy the contracts
 *  - Outputs the ABI files into bin to be later included in the distribution (for Vent and other ABI-consuming services)
 */
export async function build(srcPathOrFiles: string | string[], opts?: Partial<BuildOptions>): Promise<void> {
  const { failOnWarnings, solcVersion, binPath, basePath, burrowImportPath, abiExt } = {
    ...defaultBuildOptions,
    ...opts
  };
  const resolvedBasePath = basePath ?? (typeof srcPathOrFiles === "string" ? srcPathOrFiles : process.cwd());
  process.chdir(resolvedBasePath);
  const basePathPrefix = new RegExp("^" + path.resolve(resolvedBasePath));
  await fs.mkdir(binPath, { recursive: true });
  const solidityFiles = await getSourceFilesList(srcPathOrFiles);
  const inputDescription = inputDescriptionFromFiles(
    // solidityFiles.map((f) => path.resolve(resolvedBasePath, f.replace(basePathPrefix, ''))),
    solidityFiles
  );
  const input = encodeInput(inputDescription);
  const solc = solcCompilers[solcVersion];

  const solcOutput = solc.compile(input, { import: importLocalResolver(resolvedBasePath) });
  const output = decodeOutput(solcOutput);
  const errors = output.errors?.filter((e) => failOnWarnings || (e.severity === "error")) || [];
  if (errors.length > 0) {
    throw new Error(
      "Solidity compiler errors:\n" + formatErrors(errors)
    );
  }
  const warnings = output.errors?.filter((e) => e.severity === "warning") || [];

  if (warnings.length) {
    console.error("Solidity compiler warnings (not treated as fatal):\n" + formatErrors(warnings))
  }

  const plan = Object.keys(output.contracts).map((filename) => ({
    source: filename,
    target: filename.replace(/\.[^/.]+$/, ".abi.ts"),
    contracts: Object.entries(output.contracts[filename]).map(([name, contract]) => ({
      name,
      contract
    }))
  }));

  const binPlan = plan.flatMap((f) => {
    return f.contracts.map(({ name, contract }) => ({
      source: f.source,
      name,
      filename: path.join(binPath, path.dirname(path.resolve(f.source)).replace(basePathPrefix, ""), name + abiExt),
      abi: JSON.stringify(contract)
    }));
  });

  const dupes = findDupes(binPlan, (b) => b.filename);

  if (dupes.length) {
    const dupeDescs = dupes.map(({ key, dupes }) => ({ duplicate: key, sources: dupes.map((d) => d.source) }));
    throw Error(
      `Duplicate contract names found (these contracts will result ABI filenames that will collide since ABIs ` +
      `are flattened in '${binPath}'):\n${dupeDescs.map((d) => JSON.stringify(d)).join("\n")}`
    );
  }

  // Write the ABIs emitted for each file to the name of that file without extension. We flatten into a single
  // directory because that's what burrow deploy has always done.
  await Promise.all([
    ...binPlan.map(async ({ filename, abi }) => {
      await fs.mkdir(path.dirname(filename), { recursive: true });
      await fs.writeFile(filename, abi);
    }),
    ...plan.map(({ source, target, contracts }) =>
      fs.writeFile(
        target,
        printNodes(
          ...newFile(
            contracts.map(({ name, contract }) => getCompiled(name, contract)),
            burrowImportPath(source)
          )
        )
      )
    )
  ]);
}

function getCompiled(name: string, contract: Solidity.Contract): Compiled {
  return {
    name,
    abi: contract.abi,
    bytecode: contract.evm.bytecode.object,
    deployedBytecode: contract.evm.deployedBytecode.object,
    links: tokenizeLinks(contract.evm.bytecode.linkReferences)
  };
}

function findDupes<T>(list: T[], by: (t: T) => string): { key: string; dupes: T[] }[] {
  const grouped = list.reduce((acc, t) => {
    const k = by(t);
    if (!acc[k]) {
      acc[k] = [];
    }
    acc[k].push(t);
    return acc;
  }, {} as Record<string, T[]>);
  return Object.entries(grouped)
    .filter(([_, group]) => group.length > 1)
    .map(([key, dupes]) => ({
      key,
      dupes
    }));
}

async function getSourceFilesList(srcPathOrFiles: string | string[]): Promise<string[]> {
  if (typeof srcPathOrFiles === "string") {
    const files: string[] = [];
    for await (const f of walkDir(srcPathOrFiles)) {
      if (path.extname(f) === ".sol") {
        files.push(f);
      }
    }
    return files;
  }
  return srcPathOrFiles;
}

async function* walkDir(dir: string): AsyncGenerator<string, void, void> {
  for await (const d of await fs.opendir(dir)) {
    const entry = path.join(dir, d.name);
    if (d.isDirectory()) {
      yield* walkDir(entry);
    } else if (d.isFile()) {
      yield entry;
    }
  }
}

function formatErrors(errors: Solidity.Error[]): string {
  return errors.map((err) => err.formattedMessage || err.message).join("")
}
