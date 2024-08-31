'use client';

import { CandidateList } from "@/app/types";
import CandidateListView from "./CandidateList";
import ChatBox from "./ChatBox";
import React, { useState } from 'react';
import SystemDiagram from "./SystemDiagram";
import { useGlobalState } from "../GlobalStateProvider";

function BigBangCanvas() {
    const { objects } = useGlobalState(); // Access objects from global state

    // Filter objects by their types
    const services = objects.filter((obj) => obj.type === "service");
    const targets = objects.filter((obj) => obj.type === "target");
    const solutions = objects.filter((obj) => obj.type === "solution");
    const instances = objects.filter((obj) => obj.type === "instance");
    const sites = objects.filter((obj) => obj.type === "site");
    const catalogs = objects.filter((obj) => obj.type === "catalog");
    const campaigns = objects.filter((obj) => obj.type === "campaign");
    
     // Determine if columns should be hidden
    const isLeftColumnHidden = services.length === 0 && targets.length === 0 && sites.length === 0;
    const isRightColumnHidden = solutions.length === 0 && instances.length === 0 && catalogs.length === 0 && campaigns.length === 0;

    // Dynamically adjust column width
    const leftColumnWidth = isLeftColumnHidden ? 'hidden' : 'flex w-1/6';
    const rightColumnWidth = isRightColumnHidden ? 'hidden' : 'flex w-1/6';
    const centerColumnWidth = isLeftColumnHidden || isRightColumnHidden ? 'w-full' : 'w-4/6';

    // Determine row height dynamically
    const rowHeight = (items: any[]) => (items.length > 0 ? 'flex-1' : 'h-0');

    return (
        <div className="flex h-screen p-4">
          <div className={`flex-col ${leftColumnWidth}`}>
            <div className={`${rowHeight(services)}`}><CandidateListView name="Azure Services" type="service" /></div>
            <div className={`${rowHeight(sites)}`}><CandidateListView name="Sites" type="site" /></div>
            <div className={`${rowHeight(targets)}`}><CandidateListView name="Targets" type="target" /></div>
          </div>
          <div className={`flex flex-col ${centerColumnWidth}`}>
            <div className="flex-1"><ChatBox /></div>
            <div className="flex-1 p-4"><SystemDiagram /></div>
          </div>
          <div className={`flex-col ${rightColumnWidth}`}>
            <div className={`${rowHeight(solutions)}`}><CandidateListView name="Solutions" type="solution" /></div>
            <div className={`${rowHeight(instances)}`}><CandidateListView name="Instances" type="instance" /></div>
            <div className={`${rowHeight(campaigns)}`}><CandidateListView name="Campaigns" type="campaign" /></div>
            <div className={`${rowHeight(catalogs)}`}><CandidateListView name="Catalogs" type="catalog" /></div>
          </div>
        </div>
    );
}

export default BigBangCanvas;