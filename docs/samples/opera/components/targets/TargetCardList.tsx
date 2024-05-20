/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

import {TargetState} from '../../app/types';
import TargetCard from './TargetCard';

interface TargetCardListProps {
    targets: TargetState[];
    filter?: string;
}

function TargetCardList(props: TargetCardListProps) {
    const { targets, filter } = props;
    if (!targets) {
        return (<div>No data</div>);
    }

    const filteredTargets = filter ? targets.filter(target => target.spec.properties && target.spec.properties.scenario === filter) : targets;

    return (
        <div className='sitelist'>                                
            {filteredTargets.map((target: any) =>  {                
                return <TargetCard key={target.metadata.name} target={target} />;
            })}
        </div>
    );
}

export default TargetCardList;