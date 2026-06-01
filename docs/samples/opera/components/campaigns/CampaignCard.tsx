'use client';

import {Card, CardHeader, CardBody, CardFooter, Divider, Link, Image} from '@nextui-org/react';
import { CampaignVersionState, StageSpec, ActivationState } from '../../app/types';
import { BsArrowRightShort } from 'react-icons/bs';
import { GoDot } from 'react-icons/go';
import { Chip } from "@nextui-org/react";
import { BiPlay } from 'react-icons/bi';
import { useState } from 'react';
import {Switch} from "@nextui-org/react";
import {CgArrowLongRightC} from 'react-icons/cg';
import {LuFileJson2} from 'react-icons/lu';
import {stateToString} from "../../app/utils";

interface CampaignVersionCardProps {
    campaignversion: CampaignVersionState;
    activation?: ActivationState;
}
async function ActivateCampaignVersion(campaignversion: CampaignVersionState) {
    const response = await fetch(`/api/campaignversions/${campaignversion.id}/activate`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(campaignversion),
    });
    const data = await response.json();    
}
function CampaignVersionCard(props: CampaignVersionCardProps) {

    const { campaignversion, activation } = props;
    const stages = [];
    const stageName = campaignversion.spec.firstStage;
    const [isSelected, setIsSelected] = useState(false);

    if (stageName != "") {
        let currentStage = campaignversion.spec.stages[stageName];
        while (currentStage) {
            stages.push(currentStage);
            currentStage = campaignversion.spec.stages[currentStage.stageSelector];
        }
    }
    for (const name in campaignversion.spec.stages) {
        if (campaignversion.spec.stages.hasOwnProperty(name)) {
          const stage = campaignversion.spec.stages[name];
          if (!stages.includes(stage)) {
            stages.push(stage);
          }
        }
    }

    // get json from campaignversion with new lines
    const json = JSON.stringify(campaignversion, null, 2);    

    return (
        <Card>
            <CardHeader className="flex gap-3 justify-between">
               {campaignversion.id}
               <Switch size="sm" color="success" thumbIcon={({ isSelected, className }) =>
                    isSelected ? (
                        <LuFileJson2 className={className} />
                    ) : (
                        <CgArrowLongRightC className={className} />
                    )
                } onValueChange={setIsSelected}></Switch>
            </CardHeader>
            <Divider/>
            <CardBody>
                {!isSelected && (
                    <div className="flex">
                        {stages.map((stage: StageSpec) => (                                        
                            <div className='table-cell-no-wrap flex items-center'>
                                {stage.contexts? <Chip color="secondary">{stage.name}</Chip> : <Chip color="warning">{stage.name}</Chip>}                            
                                <BsArrowRightShort/>
                                {stage.stageSelector? '' : <GoDot/>}
                            </div>                        
                        ))}                    
                    </div>
                )}
                {isSelected && (
                    <div className="w-[600px] h-[400px]"><pre>{json}</pre></div>                
                )}
            </CardBody>
            <Divider/>
            <CardFooter  className="flex gap-3 justify-between">                
                <button className="btn btn-primary" onClick={()=>ActivateCampaignVersion(campaignversion)}><BiPlay/></button>
                {activation && (
                    <div className="flex gap-2">{` ${activation.status.statusMessage} ${activation.status.stage === '' ? '': 'stage (' + activation.status.stage + ')'}`}</div>
                )}
            </CardFooter>
        </Card>
    );
}
export default CampaignVersionCard;