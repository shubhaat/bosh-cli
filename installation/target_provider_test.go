package installation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-init/installation"

	boshlog "github.com/cloudfoundry/bosh-agent/logger"

	fakesys "github.com/cloudfoundry/bosh-agent/system/fakes"
	fakeuuid "github.com/cloudfoundry/bosh-agent/uuid/fakes"

	biconfig "github.com/cloudfoundry/bosh-init/config"
)

var _ = Describe("TargetProvider", func() {
	var (
		fakeFS                 *fakesys.FakeFileSystem
		fakeUUIDGenerator      *fakeuuid.FakeGenerator
		logger                 boshlog.Logger
		deploymentStateService biconfig.DeploymentStateService

		targetProvider TargetProvider

		configPath            = "/deployment.json"
		installationsRootPath = "/.bosh_micro/installations"
	)

	BeforeEach(func() {
		fakeFS = fakesys.NewFakeFileSystem()
		fakeUUIDGenerator = fakeuuid.NewFakeGenerator()
		logger = boshlog.NewLogger(boshlog.LevelNone)
		deploymentStateService = biconfig.NewFileSystemDeploymentStateService(
			fakeFS,
			fakeUUIDGenerator,
			logger,
			configPath,
		)
		targetProvider = NewTargetProvider(deploymentStateService, fakeUUIDGenerator, installationsRootPath)
	})

	Context("when the installation_id exists in the deployment config", func() {
		BeforeEach(func() {
			err := fakeFS.WriteFileString(configPath, `{"installation_id":"12345"}`)
			Expect(err).ToNot(HaveOccurred())
		})

		It("uses the existing installation_id & returns a target based on it", func() {
			target, err := targetProvider.NewTarget()
			Expect(err).ToNot(HaveOccurred())
			Expect(target.Path()).To(Equal("/.bosh_micro/installations/12345"))
		})

		It("does not change the saved installation_id", func() {
			_, err := targetProvider.NewTarget()
			Expect(err).ToNot(HaveOccurred())

			deploymentState, err := deploymentStateService.Load()
			Expect(err).ToNot(HaveOccurred())
			Expect(deploymentState.InstallationID).To(Equal("12345"))
		})
	})

	Context("when the installation_id does not exist in the deployment config", func() {
		BeforeEach(func() {
			err := fakeFS.WriteFileString(configPath, `{}`)
			Expect(err).ToNot(HaveOccurred())
		})

		It("generates a new installation_id & returns a target based on it", func() {
			target, err := targetProvider.NewTarget()
			Expect(err).ToNot(HaveOccurred())
			Expect(target.Path()).To(Equal("/.bosh_micro/installations/fake-uuid-1"))
		})

		It("saves the new installation_id", func() {
			_, err := targetProvider.NewTarget()
			Expect(err).ToNot(HaveOccurred())

			deploymentState, err := deploymentStateService.Load()
			Expect(err).ToNot(HaveOccurred())
			Expect(deploymentState.InstallationID).To(Equal("fake-uuid-1"))
		})
	})

	Context("when the deployment config does not exist", func() {
		BeforeEach(func() {
			err := fakeFS.RemoveAll(configPath)
			Expect(err).ToNot(HaveOccurred())
		})

		It("generates a new installation_id & returns a target based on it", func() {
			target, err := targetProvider.NewTarget()
			Expect(err).ToNot(HaveOccurred())
			Expect(target.Path()).To(Equal("/.bosh_micro/installations/fake-uuid-1"))
		})

		It("saves the new installation_id", func() {
			_, err := targetProvider.NewTarget()
			Expect(err).ToNot(HaveOccurred())

			deploymentState, err := deploymentStateService.Load()
			Expect(err).ToNot(HaveOccurred())
			Expect(deploymentState.InstallationID).To(Equal("fake-uuid-1"))
		})
	})
})
