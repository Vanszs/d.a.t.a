import { Plugin } from "@elizaos/core";
import { databaseProvider } from "./providers/ethereum/database";

export const onchainDataPlugin: Plugin = {
    name: "onchain data plugin",
    description: "Enables onchain data fetching",
    actions: [],
    providers: [databaseProvider],
    evaluators: [],
    services: [],
    clients: [],
};
