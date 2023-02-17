// Next.js API route support: https://nextjs.org/docs/api-routes/introduction
import type { NextApiRequest, NextApiResponse } from 'next'
const TwinDB = require('@proflect/node/twinDB');

export default async function handler(req, res) {  
  const name = req.query.name;
  const twin = req.query.twin;  
  const twinDb = new TwinDB();
  await twinDb.login("admin", "", "fixed");
  let p = await twinDb.getPerspective(name, twin);
  res.status(200).json(p);
}
