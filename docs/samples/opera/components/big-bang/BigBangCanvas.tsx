'use client';

import { CandidateList } from "@/app/types";
import CandidateListView from "./CandidateList";
import ChatBox from "./ChatBox";
import React, { useState } from 'react';
import SystemDiagram from "./SystemDiagram";

function BigBangCanvas() {
    const [targets, setTargets] = useState<CandidateList>({
        name: "Targets",
        candidates: [
            { name: "Azure Marketplace"}
        ]
    });
    const [services, setServices] = useState<CandidateList>({
        name: "Azure Services",
        candidates: [
            { name: "Azure Functions" },
        ]
    });
    const [solutions, setSolutions] = useState<CandidateList>({
        name: "Solutions",
        candidates: [
            { name: "Azure Marketplace"}
        ]
    });
    const [instances, setInstances] = useState<CandidateList>({
        name: "Instances",
        candidates: [
            { name: "Azure Marketplace"}
        ]
    });
    const [sites, setSites] = useState<CandidateList>({
        name: "Sites",
        candidates: [
            { name: "Azure Marketplace"}
        ]
    });
    const [catalogs, setCatalogs] = useState<CandidateList>({
        name: "Catalogs",
        candidates: [
            { name: "Azure Marketplace"}
        ]
    });
    const [campaigns, setCampaigns] = useState<CandidateList>({ 
        name: "Campaigns",
        candidates: [
            { name: "Azure Marketplace"}
        ]
    });
    const isLeftColumnHidden = services.candidates.length === 0 && targets.candidates.length === 0 && sites.candidates.length === 0;
    const isRightColumnHidden = solutions.candidates.length === 0 && instances.candidates.length === 0 && catalogs.candidates.length === 0 && campaigns.candidates.length === 0;

    const leftColumnWidth = isLeftColumnHidden ? 'hidden' : 'flex w-1/6';
    const rightColumnWidth = isRightColumnHidden ? 'hidden' : 'flex w-1/6';
    const centerColumnWidth = isLeftColumnHidden || isRightColumnHidden ? 'w-full' : 'w-4/6';

    const rowHeight = (items: CandidateList) => (items.candidates.length > 0 ? 'flex-1' : 'h-0'); // Example: Full height if content exists, or smaller height if empty

    return (
        <div className="flex h-screen p-4">
             <div className={`flex-col ${leftColumnWidth}`}>
                <div className={`${rowHeight(services)} `}><CandidateListView  {...services}/></div>
                <div className={`${rowHeight(targets)} `}><CandidateListView  {...sites} /></div>
                <div className={`${rowHeight(sites)} `}><CandidateListView  {...targets} /></div>
            </div>      
            <div className={`flex flex-col ${centerColumnWidth}`}>
                <div className="flex-1"> <ChatBox /></div>
                <div className="flex-1"><SystemDiagram /></div>
            </div>      
            <div className={`flex-col ${rightColumnWidth}`}>
                <div className={`${rowHeight(solutions)} `}><CandidateListView  {...solutions} /></div>
                <div className={`${rowHeight(instances)} `}><CandidateListView  {...instances}/></div>
                <div className={`${rowHeight(campaigns)} `}><CandidateListView  {...campaigns}/></div>
                <div className={`${rowHeight(catalogs)} `}><CandidateListView  {...catalogs}/></div>
            </div>
      </div>
    );
}

export default BigBangCanvas;