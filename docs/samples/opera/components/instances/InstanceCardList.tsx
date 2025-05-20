/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

import {InstanceState} from '../../app/types';
import InstanceCard from './InstanceCard';

interface InstanceCardListProps {
    instances: InstanceState[];
    filter?: string;
}

function InstanceCardList(props: InstanceCardListProps) {
    const { instances } = props;
    if (!instances) {
        return (<div>No data</div>);
    }

    return (
        <div className='sitelist'>            
            {instances.map((instance: any) =>  {                
                return <InstanceCard instance={instance} />;
            })}
        </div>
    );
}

export default InstanceCardList;