FROM node:22-alpine
WORKDIR /app
COPY app/package.json ./package.json
RUN npm install --omit=dev
COPY app/ ./
EXPOSE 8787
CMD ["npm", "start"]
