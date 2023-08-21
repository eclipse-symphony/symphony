'use client';

import {Card, CardHeader, CardBody, CardFooter, Divider, Link, Image} from '@nextui-org/react';
import { CampaignState, StageSpec } from '../../types';
import {Table, TableHeader, TableColumn, TableBody, TableRow, TableCell, Chip} from "@nextui-org/react";
import { BsArrowRightShort } from 'react-icons/bs';
interface CampaignCardProps {
    campaign: CampaignState;
}
function CampaignCard(props: CampaignCardProps) {
    const { campaign } = props;
    const stages = [];
    const stageName = campaign.spec.firstStage;
    if (stageName != "") {
        let currentStage = campaign.spec.stages[stageName];
        while (currentStage) {
            stages.push(currentStage);
            currentStage = campaign.spec.stages[currentStage.stageSelector];
        }
    }
    for (const name in campaign.spec.stages) {
        if (campaign.spec.stages.hasOwnProperty(name)) {
          const stage = campaign.spec.stages[name];
          if (!stages.includes(stage)) {
            stages.push(stage);
          }
        }
      }
    return (
        <Card>
            <CardHeader className="flex gap-3">
               {campaign.id}
            </CardHeader>
            <Divider/>
            <CardBody>
                <Table removeWrapper>
                    <TableHeader>
                        <TableColumn>STAGE</TableColumn>                        
                    </TableHeader>
                    <TableBody>
                        {stages.map((stage: StageSpec) => (
                            <TableRow key={stage.name}>
                                <TableCell><span  className='table-cell-no-wrap flex items-center gap-3'>{stage.name}<BsArrowRightShort/>{stage.stageSelector}</span></TableCell>                                
                            </TableRow>
                        ))}                            
                    </TableBody>
                </Table>
            </CardBody>
            {/* <Divider/>
            <CardFooter>
                    <span>MISS</span>
            </CardFooter> */}
        </Card>
    );
}
export default CampaignCard;