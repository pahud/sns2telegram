#!/usr/bin/env node
import 'source-map-support/register';
import cdk = require('@aws-cdk/core');
import { CdkS2TStack } from '../lib/cdk-stack';

const app = new cdk.App();
new CdkS2TStack(app, 'sns2telegramCdk', {
    TelegramToken: app.node.tryGetContext("telegramToken"),
    env: {
        region:  app.node.tryGetContext("region")
    }
});


