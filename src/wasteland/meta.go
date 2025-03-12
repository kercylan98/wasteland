package wasteland

import "time"

type Meta interface {
	Address() Address

	Zone() Zone

	Version() Version

	LaunchAt() Timestamp

	Gen() Gen
}

func NewMeta(address Address) Meta {
	return newMeta(address, 0, 0, time.Now().Unix(), 0)
}

func NewZoneMeta(address Address, zone Zone) Meta {
	return newMeta(address, zone, 0, time.Now().Unix(), 0)
}

func NewVersionMeta(address Address, version Version) Meta {
	return newMeta(address, 0, version, time.Now().Unix(), 0)
}

func NewZoneVersionMeta(address Address, zone Zone, version Version) Meta {
	return newMeta(address, zone, version, time.Now().Unix(), 0)
}

func newMeta(address Address, zone Zone, version Version, launchAt Timestamp, gen Gen) Meta {
	return &metaImpl{
		address:  address,
		zone:     zone,
		version:  version,
		launchAt: launchAt,
		gen:      gen,
	}
}

type metaImpl struct {
	address  Address
	zone     Zone
	version  Version
	launchAt Timestamp
	gen      Gen
}

func (p *metaImpl) Address() Address {
	return p.address
}

func (p *metaImpl) Zone() Zone {
	return p.zone
}

func (p *metaImpl) Version() Version {
	return p.version
}

func (p *metaImpl) LaunchAt() Timestamp {
	return p.launchAt
}

func (p *metaImpl) Gen() Gen {
	return p.gen
}
